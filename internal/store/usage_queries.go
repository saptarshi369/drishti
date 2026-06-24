// usage_queries.go holds the read queries for the Usage & Cost screen plus the
// quota-sample read/write. Like every store read it returns codes (never raw
// ids) and tolerates empty tables (no rows → empty slice, no error).

package store

import "github.com/saptarshi369/drishti/internal/model"

// UsageDaily returns per-day usage totals (summed across projects + models) for
// days at or after sinceDay, ascending. It reads the v_usage_daily view, so the
// agent CODE drives the filter. Label is left "" — the services layer fills the
// weekday label. est_cost_usd is summed by the view.
func (s *Store) UsageDaily(agentCode string, sinceDay int) ([]model.DailyUsage, error) {
	rows, err := s.db.Query(`
		SELECT day, input_tokens, output_tokens, cache_tokens, total_tokens, est_cost_usd
		FROM v_usage_daily
		WHERE agent = ? AND day >= ?
		ORDER BY day ASC`, agentCode, sinceDay)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []model.DailyUsage
	for rows.Next() {
		var d model.DailyUsage
		if err := rows.Scan(&d.Day, &d.InputTokens, &d.OutputTokens,
			&d.CacheTokens, &d.TotalTokens, &d.CostUSD); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// UsageByProject groups cost by the raw project key (encoded transcript dir) for
// days at or after sinceDay, ordered by cost descending. The services layer turns
// each Root into a display label and a percentage.
func (s *Store) UsageByProject(agentCode string, sinceDay int) ([]model.ProjectCost, error) {
	rows, err := s.db.Query(`
		SELECT u.project_root, SUM(u.est_cost_usd) AS cost
		FROM usage_rollup u JOIN agents a ON a.id = u.agent_id
		WHERE a.code = ? AND u.day >= ?
		GROUP BY u.project_root
		ORDER BY cost DESC`, agentCode, sinceDay)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []model.ProjectCost
	for rows.Next() {
		var p model.ProjectCost
		if err := rows.Scan(&p.Root, &p.CostUSD); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// UsageByModel groups total tokens by model for days at or after sinceDay,
// ordered by tokens descending. The services layer converts tokens into a share
// percentage and a display name.
func (s *Store) UsageByModel(agentCode string, sinceDay int) ([]model.TokensByModel, error) {
	rows, err := s.db.Query(`
		SELECT u.model, SUM(u.total_tokens) AS toks
		FROM usage_rollup u JOIN agents a ON a.id = u.agent_id
		WHERE a.code = ? AND u.day >= ?
		GROUP BY u.model
		ORDER BY toks DESC`, agentCode, sinceDay)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []model.TokensByModel
	for rows.Next() {
		var m model.TokensByModel
		if err := rows.Scan(&m.Model, &m.TotalTokens); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// InsertQuotaSample writes one quota_samples row (one window). agentID maps
// the code to its numeric id; an unknown code maps to 0. The API layer
// validates the agent code via store.AgentID and returns a typed 400 before
// calling this, so InsertQuotaSample should only ever receive known codes.
func (s *Store) InsertQuotaSample(r model.QuotaSampleRow) error {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	_, err := s.db.Exec(`
		INSERT INTO quota_samples (agent_id, ts_ms, window, used_percentage, resets_at_ms, plan, source)
		VALUES (?,?,?,?,?,?,?)`,
		agentID(r.AgentCode), r.TsMs, r.Window, r.UsedPercentage, r.ResetsAtMs, r.Plan, r.Source)
	return err
}

// LatestQuota returns the newest sample per window for an agent (from
// v_latest_quota). Empty when no samples exist — the caller renders the gated UI.
func (s *Store) LatestQuota(agentCode string) ([]model.QuotaWindowRow, error) {
	rows, err := s.db.Query(`
		SELECT window, used_percentage, resets_at_ms, ts_ms,
		       COALESCE(plan,''), COALESCE(source,'')
		FROM v_latest_quota WHERE agent = ?`, agentCode)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []model.QuotaWindowRow
	for rows.Next() {
		var w model.QuotaWindowRow
		if err := rows.Scan(&w.Window, &w.UsedPercentage, &w.ResetsAtMs, &w.TsMs, &w.Plan, &w.Source); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}
