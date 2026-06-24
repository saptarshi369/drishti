// activity_queries.go holds the read queries that power the Live Activity
// screen: the 5 event-type counters and the latest-session lookup. Like every
// store read it returns codes (never raw database ids) and tolerates empty
// tables (no rows → zero-value CounterSet or empty id, no error).

package store

import (
	"database/sql"
	"errors"

	"github.com/saptarshi369/drishti/internal/model"
)

// ActivityCounters tallies events by type for one time window.
//
// sinceMs is a Unix-millisecond lower bound (inclusive): only events whose
// ts_ms >= sinceMs are counted. Pass 0 to count all events.
//
// sessionID is an optional further filter. An empty string means "all
// sessions"; a non-empty string restricts the count to that session only.
//
// The SQL trick for the optional filter is:
//
//	(? = '' OR e.session_id = ?)
//
// When the first placeholder is ” the entire OR-branch is TRUE for every row,
// so no session filtering happens. When it is a real id the branch acts as an
// equality filter. This lets us use a single prepared statement for both
// cases — a common Go/SQLite idiom.
//
// The GROUP BY et.code query returns one row per event-type code that has at
// least one matching event; a switch maps each code to the right CounterSet
// field. Unknown codes are silently ignored (YAGNI / future-proofing).
func (s *Store) ActivityCounters(sinceMs int64, sessionID string) (model.CounterSet, error) {
	var c model.CounterSet
	// Join events → event_types so we get the human-readable code, not an int id.
	// The double-bind of sessionID (once for the '' test, once for equality) is
	// how database/sql works: each '?' consumes exactly one argument in order.
	rows, err := s.db.Query(`
		SELECT et.code, COUNT(*)
		FROM events e JOIN event_types et ON et.id = e.type_id
		WHERE e.ts_ms >= ? AND (? = '' OR e.session_id = ?)
		GROUP BY et.code`, sinceMs, sessionID, sessionID)
	if err != nil {
		return c, err
	}
	// rows.Close() is deferred even though we call rows.Err() at the end;
	// defer is the idiomatic Go safety net that also runs on early returns.
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var code string
		var n int
		if err := rows.Scan(&code, &n); err != nil {
			return c, err
		}
		// Map each event-type code to the matching CounterSet field.
		// Unknown codes (e.g. "session_start", "session_end") are intentionally
		// ignored — they are structural events, not activity counters.
		switch code {
		case "prompt":
			c.Prompts = n
		case "skill":
			c.Skills = n
		case "tool_use":
			c.Tools = n
		case "blocked":
			c.Blocked = n
		case "error":
			c.Errors = n
		}
	}
	// rows.Err() surfaces any iteration error that occurred inside rows.Next().
	// Always check it — a network/IO error mid-iteration would otherwise be lost.
	return c, rows.Err()
}

// SkillTriggers lists the active inventory skills for a project root, joined to
// their cumulative trigger counts from skill_stats. Skills with zero triggers
// are flagged Dead — they are active in the harness but have never fired, making
// them archive candidates (spec §7.6).
//
// The LEFT JOIN + COALESCE idiom: inventory_resolved drives the result set
// (every active skill appears even if skill_stats has no row for it yet). When
// no skill_stats row exists, COALESCE replaces NULL with 0 for cnt and last,
// so the scan never sees a NULL and the Dead flag is set correctly in Go.
//
// The JOIN ON clause matches all three columns of the skill_stats PRIMARY KEY:
// (agent_id, project_root, skill_name). This is critical for correctness: omitting
// project_root would cause skill_stats rows from OTHER project roots to match the
// same (agent_id, skill_name), producing duplicate result rows and corrupting counts.
// In M2 ApplyIngest always writes project_root=”, but the full 3-column join is
// enforced now to prevent cross-project contamination as more project roots appear.
//
// ORDER BY cnt DESC, r.name ASC: most-triggered skills first; ties broken
// alphabetically so output is deterministic.
//
// Empty-table tolerance: if there are no active skills the query returns zero
// rows; we return (nil, nil), not an error.
func (s *Store) SkillTriggers(projectRoot string) ([]model.SkillTrigger, error) {
	// Join inventory_resolved (active skills) to skill_stats (cumulative counts).
	// The LEFT JOIN means skills that have never fired still appear — they just
	// get NULLs from skill_stats, which COALESCE converts to 0.
	// All three PK columns are matched: agent_id anchors within the same agent,
	// project_root scopes to this project only, skill_name identifies the skill.
	rows, err := s.db.Query(`
		SELECT r.name,
		       COALESCE(ss.trigger_count_total, 0) AS cnt,
		       COALESCE(ss.last_fired_ms, 0)       AS last
		FROM inventory_resolved r
		LEFT JOIN skill_stats ss
		       ON ss.skill_name = r.name
		      AND ss.agent_id = r.agent_id
		      AND ss.project_root = r.project_root
		WHERE r.category = 'skill' AND r.effective_status = 'active' AND r.project_root = ?
		ORDER BY cnt DESC, r.name ASC`, projectRoot)
	if err != nil {
		return nil, err
	}
	// rows.Close is always deferred as a safety net; rows.Err is checked below
	// to catch any IO error that occurred mid-iteration inside rows.Next().
	defer func() { _ = rows.Close() }()
	var out []model.SkillTrigger
	for rows.Next() {
		var st model.SkillTrigger
		if err := rows.Scan(&st.Name, &st.Count, &st.LastFiredMs); err != nil {
			return nil, err
		}
		// Dead is computed in Go rather than SQL: Count==0 means the skill has
		// never fired. A LEFT JOIN with COALESCE already gave us 0 for never-fired
		// skills, so this one expression covers both "no skill_stats row" and
		// "row exists but count happens to be 0" (a future edge case).
		st.Dead = st.Count == 0
		out = append(out, st)
	}
	return out, rows.Err()
}

// SkillAnalytics lists resolved skills (active AND disabled) joined to their
// cumulative trigger counts, for the Skills Analytics screen. Unlike
// SkillTriggers (active-only, used by the Activity screen) it also returns
// disabled skills — so the screen can flag "skills you've turned off" — and it
// carries effective_status + est_context_tokens through to the analytics layer.
//
// It returns RAW rows only: value ratio and the hygiene flags are derived in
// pure Go by skills.BuildAnalytics, not in SQL (compute-on-read).
//
// Same LEFT JOIN + COALESCE idiom as SkillTriggers: inventory_resolved drives
// the result set so a never-fired skill still appears (COALESCE turns its
// missing skill_stats row into 0). All three columns of the skill_stats PRIMARY
// KEY (agent_id, project_root, skill_name) are matched, so a row from another
// project_root cannot contaminate this project's counts.
//
// ORDER BY cnt DESC, r.name ASC: most-triggered first, ties broken
// alphabetically for deterministic output. Empty result → (nil, nil), not error.
func (s *Store) SkillAnalytics(projectRoot string) ([]model.SkillStatRow, error) {
	rows, err := s.db.Query(`
		SELECT r.name,
		       r.effective_status,
		       r.est_context_tokens,
		       COALESCE(ss.trigger_count_total, 0) AS cnt,
		       COALESCE(ss.last_fired_ms, 0)       AS last
		FROM inventory_resolved r
		LEFT JOIN skill_stats ss
		       ON ss.skill_name   = r.name
		      AND ss.agent_id     = r.agent_id
		      AND ss.project_root = r.project_root
		WHERE r.category = 'skill'
		  AND r.effective_status IN ('active','disabled')
		  AND r.project_root = ?
		ORDER BY cnt DESC, r.name ASC`, projectRoot)
	if err != nil {
		return nil, err
	}
	// rows.Close is always deferred as a safety net; rows.Err is checked below
	// to catch any IO error that occurred mid-iteration inside rows.Next().
	defer func() { _ = rows.Close() }()
	var out []model.SkillStatRow
	for rows.Next() {
		var r model.SkillStatRow
		// Scan effective_status into a plain string first, then convert to the
		// typed EffectiveStatus — avoids relying on database/sql's reflection
		// fallback for custom string-kind destinations.
		var status string
		if err := rows.Scan(&r.Name, &status, &r.EstContextTokens, &r.Triggers, &r.LastFiredMs); err != nil {
			return nil, err
		}
		r.EffectiveStatus = model.EffectiveStatus(status)
		out = append(out, r)
	}
	return out, rows.Err()
}

// EventRatePerMinute returns per-minute counts of the given typeCode for the
// last `buckets` minutes ending at nowMs, oldest→newest. The result slice always
// has exactly `buckets` elements (zero-filled for minutes with no events). It
// powers the prompts/min and skills/min sparklines on the Live Activity screen.
//
// Bucketing maths (Go learner note):
//
//	A "minute index" is ts_ms / 60000 (integer division drops the sub-minute
//	remainder). Every millisecond within the same 60-second window maps to the
//	same index. SQLite computes the same division with ts_ms / 60000 in the
//	GROUP BY.
//
//	startMin = nowMs/60000 - (buckets-1) is the minute index of the OLDEST bucket
//	(e.g. buckets=3, nowMs/60000=166 → startMin=164). The SQL WHERE clause
//	filters out events before startMin*60000, so the out-of-window "old" event
//	is never returned.
//
//	Each (minuteIndex m, count n) row maps to output slot:
//	    idx = int(m - startMin)     // 0 = oldest, buckets-1 = newest
//	We range-check idx before writing so rows that fall outside [0, buckets)
//	are silently discarded (future-proofing against skewed clocks).
func (s *Store) EventRatePerMinute(typeCode string, nowMs int64, buckets int) ([]int, error) {
	// Always return a slice of the requested length, even if buckets <= 0
	// (edge-case guard: callers can safely range over an empty slice).
	out := make([]int, buckets)
	if buckets <= 0 {
		return out, nil
	}
	const minute = int64(60_000)
	// startMin is the minute index of the oldest (first) bucket.
	// nowMs/minute is the current minute index; subtracting (buckets-1) gives
	// the index that belongs in out[0].
	startMin := nowMs/minute - int64(buckets-1)
	// Query only events in the window: ts_ms >= startMin*minute (no upper bound
	// needed — events in the future belong to out[buckets-1] or are out-of-range).
	// SQLite integer division (ts_ms / 60000) mirrors Go's integer division.
	rows, err := s.db.Query(`
		SELECT e.ts_ms / 60000 AS m, COUNT(*)
		FROM events e JOIN event_types et ON et.id = e.type_id
		WHERE et.code = ? AND e.ts_ms >= ?
		GROUP BY m`, typeCode, startMin*minute)
	if err != nil {
		return nil, err
	}
	// Defer close as a safety net; rows.Err() below catches iteration errors.
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var m int64
		var n int
		if err := rows.Scan(&m, &n); err != nil {
			return nil, err
		}
		// Convert the absolute minute index to a 0-based bucket offset.
		// idx 0 = oldest (startMin), idx buckets-1 = newest (current minute).
		if idx := int(m - startMin); idx >= 0 && idx < buckets {
			out[idx] = n
		}
	}
	return out, rows.Err()
}

// EventsPage returns a newest-first page of events for the stream's scroll-back.
// typeCode "" means all types; beforeID>0 returns rows with id < beforeID (the
// keyset cursor the UI passes from the last row it holds).
//
// Keyset pagination (Go learner note):
//
//	Instead of OFFSET (which re-scans all skipped rows), we pass the id of the
//	last row the caller already has. The next page is all rows with id < beforeID,
//	ordered newest-first. This is O(1) regardless of how many rows have been
//	fetched so far — critical for a high-throughput event stream.
//
// Optional-filter SQL idiom:
//
//	(? = '' OR et.code = ?) lets typeCode="" mean "all types": when the first
//	placeholder is '' the whole OR is true for every row; when it is a real code
//	the second branch acts as an equality filter. The same double-bind trick is
//	used for beforeID=0 meaning "no cursor" via (? = 0 OR e.id < ?).
//
// Limit clamp: callers may omit the limit (0) or send an unreasonably large
// value; we normalise both to the safe default of 50 rows per page.
func (s *Store) EventsPage(typeCode string, limit int, beforeID int64) ([]model.RecentEvent, error) {
	// Clamp limit: <=0 (unset) or >200 (excessive) both default to 50 rows.
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	// The SQL selects a local `id` column (for the keyset cursor) but does NOT
	// expose it on model.RecentEvent — the stream display doesn't need the raw
	// database id; privacy and simplicity are served by omitting it.
	rows, err := s.db.Query(`
		SELECT e.id, e.ts_ms, et.code, COALESCE(e.session_id,''),
		       COALESCE(e.tool_name,''), COALESCE(e.skill_name,''), COALESCE(e.status,'')
		FROM events e JOIN event_types et ON et.id = e.type_id
		WHERE (? = '' OR et.code = ?) AND (? = 0 OR e.id < ?)
		ORDER BY e.id DESC LIMIT ?`, typeCode, typeCode, beforeID, beforeID, limit)
	if err != nil {
		return nil, err
	}
	// Defer close as a safety net — runs even if we return early on scan error.
	defer func() { _ = rows.Close() }()
	var out []model.RecentEvent
	for rows.Next() {
		var id int64 // local variable; not exposed on the model struct
		var r model.RecentEvent
		if err := rows.Scan(&id, &r.TsMs, &r.Type, &r.SessionID, &r.ToolName, &r.SkillName, &r.Status); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	// rows.Err() surfaces any IO/driver error that occurred inside rows.Next().
	// Always check it — a partial result without this check would be silently wrong.
	return out, rows.Err()
}

// LatestSession returns the most-recently-started session id and its start
// timestamp (Unix milliseconds). The UI uses this as the "current session"
// anchor for the side-by-side "session vs 24 h" view.
//
// When no sessions exist yet (empty table) it returns ("", 0, nil) rather than
// an error, which is the correct empty-table tolerance (spec §14 / failsafe).
// We detect the no-rows case with errors.Is(err, sql.ErrNoRows) — idiomatic
// Go, much safer than comparing err.Error() strings which can vary by driver.
func (s *Store) LatestSession() (string, int64, error) {
	var id string
	var started int64
	err := s.db.QueryRow(
		`SELECT session_id, COALESCE(started_ms, 0)
		 FROM sessions
		 ORDER BY started_ms DESC
		 LIMIT 1`).Scan(&id, &started)
	if err != nil {
		// sql.ErrNoRows is the sentinel value database/sql returns when
		// QueryRow finds no matching row. Return zero values, no error.
		if errors.Is(err, sql.ErrNoRows) {
			return "", 0, nil
		}
		return "", 0, err
	}
	return id, started, nil
}
