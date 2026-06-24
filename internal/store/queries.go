package store

import (
	"fmt"
	"time"

	"github.com/saptarshi369/drishti/internal/model"
)

// agentID resolves an agent code to its numeric lookup id. The slice only has
// claude(1) seeded in migration 0001 but we resolve by code so adding further
// agents later needs no query changes. Returns 0 for any unknown code.
func agentID(code string) int64 {
	if code == "claude" {
		return 1
	}
	return 0
}

// AgentID returns the numeric id for an agent code and whether that code is
// known to the daemon. The API layer uses this to reject unknown agents with a
// typed 400 BEFORE attempting any database write (a write with id=0 would trip
// the foreign-key constraint and produce a confusing 500).
//
// agentID is the single source of truth for the mapping; AgentID and
// InsertQuotaSample both delegate to it so the list is never duplicated.
//
// Example:
//
//	id, ok := store.AgentID("claude") // id=1, ok=true
//	id, ok := store.AgentID("codex")  // id=0, ok=false
func AgentID(code string) (int64, bool) {
	id := agentID(code)
	return id, id != 0
}

// eventTypeID resolves an event-type code to its lookup id.
// Codes 1–3 are seeded by migration 0001; codes 4–7 by migration 0003.
func eventTypeID(code string) int64 {
	switch code {
	case "prompt":
		return 1
	case "session_start":
		return 2
	case "session_end":
		return 3
	case "tool_use":
		return 4
	case "skill":
		return 5
	case "blocked":
		return 6
	case "error":
		return 7
	}
	return 0
}

// ListSourceFiles returns every ledger row (the reconcile ladder's input).
func (s *Store) ListSourceFiles() ([]model.SourceFile, error) {
	rows, err := s.db.Query(`
		SELECT sf.id, COALESCE(a.code,''), sf.kind, sf.abs_path, COALESCE(sf.file_id,''),
		       COALESCE(sf.size,0), COALESCE(sf.mtime_ms,0), COALESCE(sf.head_hash,''),
		       sf.last_offset, sf.last_line, sf.state, sf.error_count
		FROM source_files sf LEFT JOIN agents a ON a.id = sf.agent_id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []model.SourceFile
	for rows.Next() {
		var f model.SourceFile
		if err := rows.Scan(&f.ID, &f.AgentCode, &f.Kind, &f.AbsPath, &f.FileID,
			&f.Size, &f.MtimeMs, &f.HeadHash, &f.LastOffset, &f.LastLine, &f.State, &f.ErrorCount); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// UpsertSourceFile inserts or updates a ledger row keyed by abs_path and
// returns its id.
func (s *Store) UpsertSourceFile(sf model.SourceFile) (int64, error) {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	now := time.Now().UnixMilli()
	_, err := s.db.Exec(`
		INSERT INTO source_files (agent_id, kind, abs_path, file_id, size, mtime_ms,
		    head_hash, last_offset, last_line, state, first_seen_ms, last_read_ms, error_count)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(abs_path) DO UPDATE SET
		    file_id=excluded.file_id, size=excluded.size, mtime_ms=excluded.mtime_ms,
		    head_hash=excluded.head_hash, last_offset=excluded.last_offset,
		    last_line=excluded.last_line, state=excluded.state, last_read_ms=excluded.last_read_ms`,
		agentID(sf.AgentCode), sf.Kind, sf.AbsPath, sf.FileID, sf.Size, sf.MtimeMs,
		sf.HeadHash, sf.LastOffset, sf.LastLine, sf.State, now, now, sf.ErrorCount)
	if err != nil {
		return 0, err
	}
	var id int64
	if err := s.db.QueryRow("SELECT id FROM source_files WHERE abs_path=?", sf.AbsPath).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// IngestBatch is one atomic unit of ingestion: the rows parsed from a file
// region plus the new ledger offset to commit alongside them.
type IngestBatch struct {
	SourceFileID int64
	// ProjectRoot is the per-transcript project key (the encoded ~/.claude/projects
	// directory name) used to attribute usage_rollup rows. "" = unknown/user-global.
	ProjectRoot string
	Events      []model.Event
	Deltas      []model.SessionDelta
	NewOffset   int64
	NewLine     int64
	ReadMs      int64
}

// ApplyIngest writes one batch in a SINGLE transaction: dedupe-insert events,
// upsert session counters, fold the daily usage_rollup, and advance the ledger
// offset. The offset can never claim more than the rows actually stored, and
// re-presenting a line is a no-op (ON CONFLICT DO NOTHING). Returns the number
// of NEW event rows inserted.
func (s *Store) ApplyIngest(b IngestBatch) (int, error) {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	inserted := 0
	// rollupKey identifies one usage_rollup row touched by this batch. We collect
	// the distinct keys so that — after the token totals are folded — we can stamp
	// est_cost_usd on exactly those rows (cheap), instead of the read path later
	// rewriting the whole table on every broadcast.
	type rollupKey struct {
		day   int
		model string
	}
	touched := map[rollupKey]struct{}{}
	for _, e := range b.Events {
		// Write the Module 2 event columns: tool_name, skill_name, status.
		// nullIfEmpty stores "" as SQL NULL (keeps partial indexes lean).
		res, err := tx.Exec(`
			INSERT INTO events (agent_id, type_id, source_id, session_id, ts_ms,
			    tool_name, skill_name, status, dedupe_key)
			VALUES (?,?,1,?,?,?,?,?,?) ON CONFLICT(dedupe_key) DO NOTHING`,
			agentID(e.AgentCode), eventTypeID(e.TypeCode), e.SessionID, e.TsMs,
			nullIfEmpty(e.ToolName), nullIfEmpty(e.SkillName), nullIfEmpty(e.Status), e.DedupeKey)
		if err != nil {
			return 0, fmt.Errorf("insert event: %w", err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			inserted++
			if e.TypeCode == "skill" && e.SkillName != "" {
				// Upsert the cumulative skill counters. MIN/MAX via the excluded
				// pseudo-table let SQLite do the comparison without an extra read:
				//   first_fired_ms: stays at the earliest timestamp ever seen for
				//                   this skill (set once on first insert, never raised).
				//   last_fired_ms:  advances to the latest timestamp ever seen
				//                   (handles events arriving out-of-order correctly).
				if _, err := tx.Exec(`
					INSERT INTO skill_stats (agent_id, project_root, skill_name,
					    trigger_count_total, first_fired_ms, last_fired_ms)
					VALUES (1,'',?,1,?,?)
					ON CONFLICT(agent_id, project_root, skill_name) DO UPDATE SET
					    trigger_count_total = skill_stats.trigger_count_total + 1,
					    first_fired_ms = MIN(skill_stats.first_fired_ms, excluded.first_fired_ms),
					    last_fired_ms  = MAX(skill_stats.last_fired_ms,  excluded.last_fired_ms)`,
					e.SkillName, e.TsMs, e.TsMs); err != nil {
					return 0, fmt.Errorf("fold skill_stats: %w", err)
				}
			}
		}
	}

	for _, d := range b.Deltas {
		if _, err := tx.Exec(`
			INSERT INTO sessions (agent_id, session_id, model, started_ms, prompt_count,
			    input_tokens, output_tokens, cache_tokens, est_cost_usd)
			VALUES (1,?,?,?,?,?,?,?,0)
			ON CONFLICT(agent_id, session_id) DO UPDATE SET
			    model=excluded.model,
			    started_ms=MIN(sessions.started_ms, excluded.started_ms),
			    prompt_count=sessions.prompt_count+excluded.prompt_count,
			    input_tokens=sessions.input_tokens+excluded.input_tokens,
			    output_tokens=sessions.output_tokens+excluded.output_tokens,
			    cache_tokens=sessions.cache_tokens+excluded.cache_tokens`,
			d.SessionID, d.Model, d.StartedMs, d.PromptCount,
			d.InputTokens, d.OutputTokens, d.CacheTokens); err != nil {
			return 0, fmt.Errorf("upsert session: %w", err)
		}

		total := d.InputTokens + d.OutputTokens + d.CacheTokens
		if _, err := tx.Exec(`
			INSERT INTO usage_rollup (agent_id, day, project_root, model,
			    input_tokens, output_tokens, cache_tokens, total_tokens, prompt_count)
			VALUES (1,?, ?, ?, ?,?,?,?,?)
			ON CONFLICT(agent_id, day, project_root, model) DO UPDATE SET
			    input_tokens=usage_rollup.input_tokens+excluded.input_tokens,
			    output_tokens=usage_rollup.output_tokens+excluded.output_tokens,
			    cache_tokens=usage_rollup.cache_tokens+excluded.cache_tokens,
			    total_tokens=usage_rollup.total_tokens+excluded.total_tokens,
			    prompt_count=usage_rollup.prompt_count+excluded.prompt_count`,
			d.Day, b.ProjectRoot, d.Model, d.InputTokens, d.OutputTokens, d.CacheTokens, total, d.PromptCount); err != nil {
			return 0, fmt.Errorf("fold rollup: %w", err)
		}
		touched[rollupKey{day: d.Day, model: d.Model}] = struct{}{}
	}

	// Stamp est_cost_usd on the rows this batch folded, computed from each row's
	// now-current totals. Doing it here — once per touched row, inside the same
	// transaction — keeps cost correct without the read path ever rewriting the
	// whole table. cache tokens are predominantly reads, priced at the read rate
	// (mirrors BackfillRollupCost). A nil costFn (no pricing wired) skips this.
	if s.costFn != nil {
		for k := range touched {
			var id, in, out, cache int64
			if err := tx.QueryRow(`
				SELECT id, input_tokens, output_tokens, cache_tokens FROM usage_rollup
				WHERE agent_id=1 AND day=? AND project_root=? AND model=?`,
				k.day, b.ProjectRoot, k.model).Scan(&id, &in, &out, &cache); err != nil {
				return 0, fmt.Errorf("read folded rollup for cost: %w", err)
			}
			cost := s.costFn(k.model, in, out, cache, 0)
			if _, err := tx.Exec(`UPDATE usage_rollup SET est_cost_usd=? WHERE id=?`, cost, id); err != nil {
				return 0, fmt.Errorf("stamp rollup cost: %w", err)
			}
		}
	}

	if b.SourceFileID != 0 {
		if _, err := tx.Exec(`
			UPDATE source_files SET last_offset=?, last_line=?, last_read_ms=?, state='active'
			WHERE id=?`, b.NewOffset, b.NewLine, b.ReadMs, b.SourceFileID); err != nil {
			return 0, fmt.Errorf("advance ledger: %w", err)
		}
	}
	return inserted, tx.Commit()
}

// RecentEvents returns the newest events for the live ticker, joined to codes.
// The three optional columns (tool_name, skill_name, status) are stored as SQL
// NULL in the events table when not applicable (e.g. prompt rows). COALESCE
// converts NULL to "" so rows.Scan always receives a string, never a nil
// interface — a common Go/SQLite idiom to avoid sql.NullString boilerplate.
// The omitempty JSON tags on RecentEvent mean empty fields are stripped from
// the API response, keeping prompt-row payloads compact.
func (s *Store) RecentEvents(limit int) ([]model.RecentEvent, error) {
	rows, err := s.db.Query(`
		SELECT e.id, e.ts_ms, et.code, COALESCE(e.session_id,''),
		       COALESCE(e.tool_name,''), COALESCE(e.skill_name,''), COALESCE(e.status,'')
		FROM events e JOIN event_types et ON et.id = e.type_id
		ORDER BY e.id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []model.RecentEvent
	for rows.Next() {
		var r model.RecentEvent
		if err := rows.Scan(&r.ID, &r.TsMs, &r.Type, &r.SessionID, &r.ToolName, &r.SkillName, &r.Status); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// nullIfEmpty returns nil for an empty string so optional event columns (tool_name,
// skill_name, status) are stored as SQL NULL rather than "". This keeps partial
// indexes lean (only rows where the column IS NOT NULL are indexed) and makes
// aggregation queries cleaner — a Go learner tip: any interface{} / any holding
// nil is the database/sql signal for SQL NULL on write.
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// BackfillRollupCost recomputes est_cost_usd for every rollup row from its own
// token totals using the supplied pricing function. It SETS (not adds) the
// value, so it is idempotent and safe to call before every KPI read. The store
// stays pricing-agnostic: the cost function is injected by the services layer.
func (s *Store) BackfillRollupCost(costFn func(model string, in, out, cacheRead, cacheWrite int64) float64) error {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	rows, err := s.db.Query(`SELECT id, model, input_tokens, output_tokens, cache_tokens FROM usage_rollup`)
	if err != nil {
		return err
	}
	type upd struct {
		id   int64
		cost float64
	}
	var updates []upd
	for rows.Next() {
		var id, in, out, cache int64
		var modelName string
		if err := rows.Scan(&id, &modelName, &in, &out, &cache); err != nil {
			_ = rows.Close()
			return err
		}
		// cache tokens here are predominantly reads; price them at the read rate.
		updates = append(updates, upd{id, costFn(modelName, in, out, cache, 0)})
	}
	_ = rows.Close()
	for _, u := range updates {
		if _, err := s.db.Exec(`UPDATE usage_rollup SET est_cost_usd=? WHERE id=?`, u.cost, u.id); err != nil {
			return err
		}
	}
	return nil
}

// OverviewKPIs reads today's headline numbers from v_overview_kpis (claude row).
func (s *Store) OverviewKPIs() (model.OverviewKPIs, error) {
	var k model.OverviewKPIs
	err := s.db.QueryRow(`
		SELECT prompts_today, spend_today_usd, input_tokens, output_tokens, cache_tokens
		FROM v_overview_kpis WHERE agent='claude'`).
		Scan(&k.PromptsToday, &k.SpendTodayUSD, &k.InputTokens, &k.OutputTokens, &k.CacheTokens)
	return k, err
}
