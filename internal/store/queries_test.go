package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/saptarshi369/drishti/internal/model"
)

func todayYYYYMMDD() int {
	n := time.Now()
	return n.Year()*10000 + int(n.Month())*100 + n.Day()
}

// TestProjectKeyFiltersEventsAndUsage proves the Overview root filter: events and
// usage carry a project key (the encoded Claude project dir), and the read queries
// filter by it. An empty key means "All" (no filter). This is what lets the
// top-bar dropdown scope spend/prompts AND the live-activity feed to one folder.
func TestProjectKeyFiltersEventsAndUsage(t *testing.T) {
	p := filepath.Join(t.TempDir(), "drishti.db")
	s, _ := Open(p)
	defer s.Close()
	s.SetCostFn(func(_ string, in, _, _, _ int64) float64 { return float64(in) }) // $1/input token
	day := todayYYYYMMDD()
	now := time.Now().UnixMilli()
	sf, _ := s.UpsertSourceFile(model.SourceFile{AgentCode: "claude", Kind: "transcript", AbsPath: "/t.jsonl", State: "active"})

	ingest := func(key, sess, dedupe string, in int64) {
		s.ApplyIngest(IngestBatch{
			SourceFileID: sf,
			ProjectRoot:  key,
			Events:       []model.Event{{AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript", SessionID: sess, TsMs: now, DedupeKey: dedupe}},
			Deltas:       []model.SessionDelta{{SessionID: sess, Model: "claude-opus-4-8", Day: day, InputTokens: in, PromptCount: 1, StartedMs: 1}},
		})
	}
	ingest("-proj-a", "sa", "a", 10)
	ingest("-proj-b", "sb", "b", 5)

	// RecentEvents: key filters; "" returns all.
	if got, _ := s.RecentEvents(10, "-proj-a"); len(got) != 1 {
		t.Errorf("RecentEvents(-proj-a) = %d events, want 1", len(got))
	}
	if got, _ := s.RecentEvents(10, ""); len(got) != 2 {
		t.Errorf("RecentEvents(all) = %d events, want 2", len(got))
	}

	// OverviewKPIs: spend + prompts scoped by key.
	ka, _ := s.OverviewKPIs("-proj-a")
	if ka.PromptsToday != 1 || ka.SpendTodayUSD != 10 {
		t.Errorf("OverviewKPIs(-proj-a) = %+v, want prompts=1 spend=10", ka)
	}
	kAll, _ := s.OverviewKPIs("")
	if kAll.PromptsToday != 2 || kAll.SpendTodayUSD != 15 {
		t.Errorf("OverviewKPIs(all) = %+v, want prompts=2 spend=15", kAll)
	}

	// ActivityCounters + EventRatePerMinute: event tallies scoped by key.
	ca, _ := s.ActivityCounters(0, "", "-proj-a")
	if ca.Prompts != 1 {
		t.Errorf("ActivityCounters(-proj-a).Prompts = %d, want 1", ca.Prompts)
	}
	cAll, _ := s.ActivityCounters(0, "", "")
	if cAll.Prompts != 2 {
		t.Errorf("ActivityCounters(all).Prompts = %d, want 2", cAll.Prompts)
	}
	ra, _ := s.EventRatePerMinute("prompt", now, 5, "-proj-a")
	sum := 0
	for _, v := range ra {
		sum += v
	}
	if sum != 1 {
		t.Errorf("EventRatePerMinute(-proj-a) sum = %d, want 1", sum)
	}
}

func TestApplyIngestInsertsAndIsIdempotent(t *testing.T) {
	p := filepath.Join(t.TempDir(), "drishti.db")
	s, _ := Open(p)
	defer s.Close()

	sfID, err := s.UpsertSourceFile(model.SourceFile{
		AgentCode: "claude", Kind: "transcript", AbsPath: "/x/a.jsonl", State: "active",
	})
	if err != nil {
		t.Fatal(err)
	}

	batch := IngestBatch{
		SourceFileID: sfID,
		Events: []model.Event{{
			AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript",
			SessionID: "s1", TsMs: 1, DedupeKey: "claude|s1|h1",
		}},
		Deltas: []model.SessionDelta{{
			SessionID: "s1", Model: "claude-opus-4-8", Day: 20260621,
			InputTokens: 100, OutputTokens: 20, PromptCount: 1, StartedMs: 1,
		}},
		NewOffset: 50, NewLine: 1, ReadMs: 5,
	}
	n1, err := s.ApplyIngest(batch)
	if err != nil || n1 != 1 {
		t.Fatalf("first apply inserted=%d err=%v, want 1", n1, err)
	}
	n2, err := s.ApplyIngest(batch)
	if err != nil || n2 != 0 {
		t.Fatalf("replay inserted=%d err=%v, want 0", n2, err)
	}

	files, _ := s.ListSourceFiles()
	if files[0].LastOffset != 50 {
		t.Errorf("offset = %d, want 50", files[0].LastOffset)
	}
}

func TestOverviewKPIsReadsRollup(t *testing.T) {
	p := filepath.Join(t.TempDir(), "drishti.db")
	s, _ := Open(p)
	defer s.Close()

	day := todayYYYYMMDD()
	sfID, _ := s.UpsertSourceFile(model.SourceFile{AgentCode: "claude", Kind: "transcript", AbsPath: "/x/b.jsonl", State: "active"})
	_, err := s.ApplyIngest(IngestBatch{
		SourceFileID: sfID,
		Deltas: []model.SessionDelta{{
			SessionID: "s9", Model: "claude-opus-4-8", Day: day,
			InputTokens: 1000, OutputTokens: 200, PromptCount: 2, StartedMs: 1,
		}},
		Events:    []model.Event{{AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript", SessionID: "s9", TsMs: 1, DedupeKey: "claude|s9|p1"}},
		NewOffset: 10, NewLine: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	k, err := s.OverviewKPIs("")
	if err != nil {
		t.Fatal(err)
	}
	if k.PromptsToday != 2 {
		t.Errorf("prompts_today = %d, want 2", k.PromptsToday)
	}
	if k.InputTokens != 1000 {
		t.Errorf("input_tokens = %d, want 1000", k.InputTokens)
	}
}

func TestRecentEventsNewestFirst(t *testing.T) {
	p := filepath.Join(t.TempDir(), "drishti.db")
	s, _ := Open(p)
	defer s.Close()
	sf, _ := s.UpsertSourceFile(model.SourceFile{AgentCode: "claude", Kind: "transcript", AbsPath: "/r.jsonl", State: "active"})
	s.ApplyIngest(IngestBatch{
		SourceFileID: sf,
		Events: []model.Event{
			{AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript", SessionID: "s1", TsMs: 1, DedupeKey: "k1"},
			{AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript", SessionID: "s1", TsMs: 2, DedupeKey: "k2"},
		},
		NewOffset: 1,
	})
	got, err := s.RecentEvents(10, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("recent = %d, want 2", len(got))
	}
	if got[0].Type != "prompt" || got[0].TsMs != 2 {
		t.Errorf("newest-first broken: %+v", got[0])
	}
}

// TestRecentEventsExposeUniqueIDs proves each RecentEvent carries a distinct DB
// id so the live-stream {#each} can key on it. Two events that share
// ts_ms+type+session_id — a real collision that crashed the UI with Svelte's
// each_key_duplicate — must still come back with different ids.
func TestRecentEventsExposeUniqueIDs(t *testing.T) {
	p := filepath.Join(t.TempDir(), "drishti.db")
	s, _ := Open(p)
	defer s.Close()
	sf, _ := s.UpsertSourceFile(model.SourceFile{AgentCode: "claude", Kind: "transcript", AbsPath: "/r.jsonl", State: "active"})
	s.ApplyIngest(IngestBatch{
		SourceFileID: sf,
		Events: []model.Event{
			{AgentCode: "claude", TypeCode: "tool_use", SourceCode: "transcript", SessionID: "s1", TsMs: 5, ToolName: "Bash", DedupeKey: "a"},
			{AgentCode: "claude", TypeCode: "tool_use", SourceCode: "transcript", SessionID: "s1", TsMs: 5, ToolName: "Bash", DedupeKey: "b"},
		},
		NewOffset: 1,
	})
	got, err := s.RecentEvents(10, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("recent = %d, want 2", len(got))
	}
	if got[0].ID == 0 || got[1].ID == 0 {
		t.Fatalf("ids not populated: %+v", got)
	}
	if got[0].ID == got[1].ID {
		t.Errorf("ids not unique: both got %d", got[0].ID)
	}
}

// TestApplyIngestFoldsSkillStats verifies that ApplyIngest:
//  1. persists the tool_name column on a tool_use event, and
//  2. folds skill events into skill_stats (trigger_count_total increments,
//     last_fired_ms tracks the MAX timestamp) within the same transaction.
//
// Two "deploy" skill events are submitted; we expect count=2 and last=2000.
func TestApplyIngestFoldsSkillStats(t *testing.T) {
	st := tempStore(t)
	batch := IngestBatch{
		Events: []model.Event{
			{AgentCode: "claude", TypeCode: "skill", SourceCode: "transcript", SessionID: "s1", TsMs: 1000, SkillName: "deploy", DedupeKey: "claude|s1|tu1"},
			{AgentCode: "claude", TypeCode: "skill", SourceCode: "transcript", SessionID: "s1", TsMs: 2000, SkillName: "deploy", DedupeKey: "claude|s1|tu2"},
			{AgentCode: "claude", TypeCode: "tool_use", SourceCode: "transcript", SessionID: "s1", TsMs: 1500, ToolName: "Bash", DedupeKey: "claude|s1|tu3"},
		},
		ReadMs: 9,
	}
	if _, err := st.ApplyIngest(batch); err != nil {
		t.Fatal(err)
	}
	var cnt int
	var last int64
	if err := st.db.QueryRow(
		`SELECT trigger_count_total, last_fired_ms FROM skill_stats WHERE skill_name='deploy'`).Scan(&cnt, &last); err != nil {
		t.Fatal(err)
	}
	if cnt != 2 || last != 2000 {
		t.Fatalf("skill_stats deploy: count=%d last=%d (want 2, 2000)", cnt, last)
	}
	var first int64
	if err := st.db.QueryRow(
		`SELECT first_fired_ms FROM skill_stats WHERE skill_name='deploy'`).Scan(&first); err != nil {
		t.Fatal(err)
	}
	if first != 1000 {
		t.Fatalf("skill_stats deploy: first_fired_ms=%d (want 1000)", first)
	}
	var tool string
	if err := st.db.QueryRow(
		`SELECT tool_name FROM events WHERE dedupe_key='claude|s1|tu3'`).Scan(&tool); err != nil || tool != "Bash" {
		t.Fatalf("tool_name=%q err=%v", tool, err)
	}
}

// TestApplyIngestUsesBatchProjectRoot verifies the usage_rollup fold attributes
// rows to IngestBatch.ProjectRoot instead of the old hardcoded ”.
func TestApplyIngestUsesBatchProjectRoot(t *testing.T) {
	p := filepath.Join(t.TempDir(), "drishti.db")
	s, _ := Open(p)
	defer s.Close()

	_, err := s.ApplyIngest(IngestBatch{
		ProjectRoot: "-Users-me-dev-myapp",
		Deltas: []model.SessionDelta{{
			SessionID: "s1", Model: "claude-opus-4-8", Day: 20260622,
			InputTokens: 100, OutputTokens: 20, PromptCount: 1, StartedMs: 1,
		}},
		Events: []model.Event{{
			AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript",
			SessionID: "s1", TsMs: 1, DedupeKey: "k1",
		}},
		NewOffset: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	var root string
	if err := s.DB().QueryRow(
		`SELECT project_root FROM usage_rollup WHERE day=20260622`).Scan(&root); err != nil {
		t.Fatal(err)
	}
	if root != "-Users-me-dev-myapp" {
		t.Fatalf("project_root = %q, want -Users-me-dev-myapp", root)
	}
}

// TestAgentID verifies the exported AgentID helper returns the correct id and
// known flag for both a recognised code ("claude") and an unknown one ("codex").
// This guards the API-layer 400 path that relies on AgentID to pre-validate the
// agent before any database write.
func TestAgentID(t *testing.T) {
	if id, ok := AgentID("claude"); id != 1 || !ok {
		t.Errorf("AgentID(claude) = (%d,%v), want (1,true)", id, ok)
	}
	if id, ok := AgentID("codex"); id != 0 || ok {
		t.Errorf("AgentID(codex) = (%d,%v), want (0,false)", id, ok)
	}
}

// TestApplyIngestComputesRollupCostFromInjectedFn proves the perf fix: once a
// pricing function is injected via SetCostFn, ApplyIngest itself sets
// est_cost_usd on the rollup row — so the read/broadcast path never has to call
// BackfillRollupCost. The cost is recomputed from the row's *folded* total, so a
// second batch on the same day/model updates the cost to match the new total
// (SET, not naive-add): in=10 then in=5 → total in=15 → cost 15 with a $1/input
// pricing fn. No BackfillRollupCost call appears in this test on purpose.
func TestApplyIngestComputesRollupCostFromInjectedFn(t *testing.T) {
	p := filepath.Join(t.TempDir(), "drishti.db")
	s, _ := Open(p)
	defer s.Close()
	// Pricing fn = $1 per input token, so cost tracks the folded input total and
	// we can tell "set from current total" apart from "added per batch".
	s.SetCostFn(func(_ string, in, _, _, _ int64) float64 { return float64(in) })
	day := todayYYYYMMDD()
	sf, _ := s.UpsertSourceFile(model.SourceFile{AgentCode: "claude", Kind: "transcript", AbsPath: "/c.jsonl", State: "active"})

	s.ApplyIngest(IngestBatch{
		SourceFileID: sf,
		Deltas:       []model.SessionDelta{{SessionID: "s1", Model: "m", Day: day, InputTokens: 10, StartedMs: 1}},
		NewOffset:    1,
	})
	var got float64
	s.DB().QueryRow(`SELECT est_cost_usd FROM usage_rollup WHERE day=?`, day).Scan(&got)
	if got != 10.0 {
		t.Fatalf("after first ingest est_cost_usd = %v, want 10.0 (cost set at ingest)", got)
	}

	// Fold a second batch into the SAME rollup row (same day+model). Total input
	// becomes 15, so the recomputed cost must be 15 — not left at 10, not 10+5
	// added blindly without reading the folded total.
	s.ApplyIngest(IngestBatch{
		SourceFileID: sf,
		Deltas:       []model.SessionDelta{{SessionID: "s2", Model: "m", Day: day, InputTokens: 5, StartedMs: 2}},
		NewOffset:    2,
	})
	s.DB().QueryRow(`SELECT est_cost_usd FROM usage_rollup WHERE day=?`, day).Scan(&got)
	if got != 15.0 {
		t.Errorf("after second ingest est_cost_usd = %v, want 15.0 (recomputed from folded total)", got)
	}
}

func TestBackfillRollupCostIsIdempotent(t *testing.T) {
	p := filepath.Join(t.TempDir(), "drishti.db")
	s, _ := Open(p)
	defer s.Close()
	day := todayYYYYMMDD()
	sf, _ := s.UpsertSourceFile(model.SourceFile{AgentCode: "claude", Kind: "transcript", AbsPath: "/c.jsonl", State: "active"})
	s.ApplyIngest(IngestBatch{
		SourceFileID: sf,
		Deltas:       []model.SessionDelta{{SessionID: "s1", Model: "m", Day: day, InputTokens: 10, StartedMs: 1}},
		NewOffset:    1,
	})
	// costFn returns a fixed $2 regardless of inputs; SET (not add) → idempotent.
	cost := func(string, int64, int64, int64, int64) float64 { return 2.0 }
	if err := s.BackfillRollupCost(cost); err != nil {
		t.Fatal(err)
	}
	if err := s.BackfillRollupCost(cost); err != nil {
		t.Fatal(err)
	}
	var got float64
	s.DB().QueryRow(`SELECT est_cost_usd FROM usage_rollup WHERE day=?`, day).Scan(&got)
	if got != 2.0 {
		t.Errorf("est_cost_usd = %v, want 2.0 (set, not accumulated)", got)
	}
}
