package store

import (
	"path/filepath"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

// TestMigration0004Schema verifies migration 0004 creates the quota_samples
// table and the two new views, and that v_usage_daily sums usage_rollup rows
// per (agent, day) returning the agent CODE (never a raw id).
func TestMigration0004Schema(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "drishti.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// quota_samples accepts a row and the index exists (insert must not error).
	if _, err := s.DB().Exec(
		`INSERT INTO quota_samples (agent_id, ts_ms, window, used_percentage, resets_at_ms, plan, source)
		 VALUES (1, 1000, 'five_hour', 41.0, 2000, 'max', 'statusline')`); err != nil {
		t.Fatalf("insert quota_samples: %v", err)
	}

	// v_latest_quota returns that row with the agent code joined in.
	var agent, window string
	if err := s.DB().QueryRow(
		`SELECT agent, window FROM v_latest_quota WHERE window='five_hour'`).
		Scan(&agent, &window); err != nil {
		t.Fatalf("v_latest_quota: %v", err)
	}
	if agent != "claude" || window != "five_hour" {
		t.Fatalf("v_latest_quota = %q/%q, want claude/five_hour", agent, window)
	}

	// Seed two usage_rollup rows for the same (agent, day) but different models;
	// v_usage_daily must sum them into one row.
	for _, m := range []string{"claude-opus-4-8", "claude-sonnet-4-6"} {
		if _, err := s.DB().Exec(
			`INSERT INTO usage_rollup (agent_id, day, project_root, model, input_tokens, total_tokens)
			 VALUES (1, 20260622, '', ?, 100, 100)`, m); err != nil {
			t.Fatalf("seed rollup: %v", err)
		}
	}
	var input, total int64
	if err := s.DB().QueryRow(
		`SELECT input_tokens, total_tokens FROM v_usage_daily WHERE day=20260622`).
		Scan(&input, &total); err != nil {
		t.Fatalf("v_usage_daily: %v", err)
	}
	if input != 200 || total != 200 {
		t.Fatalf("v_usage_daily sums = %d/%d, want 200/200", input, total)
	}
}

// TestUsageDailyWindow verifies UsageDaily returns per-day summed rows at/after
// sinceDay, ascending, with token classes and cost populated.
func TestUsageDailyWindow(t *testing.T) {
	s, _ := Open(filepath.Join(t.TempDir(), "drishti.db"))
	defer s.Close()
	seed := func(day int, root, model string, in, out, cache int64, cost float64) {
		total := in + out + cache
		if _, err := s.DB().Exec(
			`INSERT INTO usage_rollup (agent_id, day, project_root, model,
			   input_tokens, output_tokens, cache_tokens, total_tokens, est_cost_usd)
			 VALUES (1,?,?,?,?,?,?,?,?)`, day, root, model, in, out, cache, total, cost); err != nil {
			t.Fatal(err)
		}
	}
	seed(20260620, "-a", "claude-opus-4-8", 100, 10, 5, 1.5)
	seed(20260622, "-a", "claude-opus-4-8", 200, 20, 0, 2.0)
	seed(20260622, "-b", "claude-sonnet-4-6", 50, 5, 0, 0.5) // same day, other project

	rows, err := s.UsageDaily("claude", 20260621)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 { // only 20260622 is >= sinceDay
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if rows[0].Day != 20260622 || rows[0].InputTokens != 250 || rows[0].CostUSD != 2.5 {
		t.Fatalf("row = %+v, want day 20260622 input 250 cost 2.5", rows[0])
	}
}

// TestUsageByProjectAndModel verifies the two grouped breakdowns sum correctly.
func TestUsageByProjectAndModel(t *testing.T) {
	s, _ := Open(filepath.Join(t.TempDir(), "drishti.db"))
	defer s.Close()
	mk := func(root, model string, total int64, cost float64) {
		if _, err := s.DB().Exec(
			`INSERT INTO usage_rollup (agent_id, day, project_root, model, total_tokens, est_cost_usd)
			 VALUES (1, 20260622, ?, ?, ?, ?)`, root, model, total, cost); err != nil {
			t.Fatal(err)
		}
	}
	mk("-myapp", "claude-opus-4-8", 100, 12.0)
	mk("-myapp", "claude-sonnet-4-6", 50, 1.0)
	mk("-website", "claude-opus-4-8", 30, 3.0)

	proj, err := s.UsageByProject("claude", 20260601)
	if err != nil {
		t.Fatal(err)
	}
	// Sorted by cost desc inside the query: -myapp (13.0) then -website (3.0).
	if len(proj) != 2 || proj[0].Root != "-myapp" || proj[0].CostUSD != 13.0 {
		t.Fatalf("byProject = %+v", proj)
	}
	mdl, err := s.UsageByModel("claude", 20260601)
	if err != nil {
		t.Fatal(err)
	}
	// opus total = 130, sonnet = 50; sorted desc by tokens.
	if len(mdl) != 2 || mdl[0].Model != "claude-opus-4-8" || mdl[0].TotalTokens != 130 {
		t.Fatalf("byModel = %+v", mdl)
	}
}

// TestQuotaInsertAndLatest verifies InsertQuotaSample writes one row per window
// and LatestQuota returns only the newest sample per window.
func TestQuotaInsertAndLatest(t *testing.T) {
	s, _ := Open(filepath.Join(t.TempDir(), "drishti.db"))
	defer s.Close()
	rows := []model.QuotaSampleRow{
		{AgentCode: "claude", Window: "five_hour", UsedPercentage: 10, ResetsAtMs: 5, TsMs: 1000, Plan: "max", Source: "statusline"},
		{AgentCode: "claude", Window: "five_hour", UsedPercentage: 41, ResetsAtMs: 9, TsMs: 2000, Plan: "max", Source: "statusline"},
		{AgentCode: "claude", Window: "seven_day", UsedPercentage: 12, ResetsAtMs: 7, TsMs: 1500, Plan: "max", Source: "statusline"},
	}
	for _, r := range rows {
		if err := s.InsertQuotaSample(r); err != nil {
			t.Fatal(err)
		}
	}
	latest, err := s.LatestQuota("claude")
	if err != nil {
		t.Fatal(err)
	}
	if len(latest) != 2 {
		t.Fatalf("latest windows = %d, want 2", len(latest))
	}
	for _, w := range latest {
		if w.Window == "five_hour" && w.UsedPercentage != 41 {
			t.Fatalf("five_hour latest = %v, want 41 (newest)", w.UsedPercentage)
		}
	}
}
