package store

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

// TestEventTypeIDActivityCodes verifies the 4 new event-type codes added in
// migration 0003 resolve to IDs 4–7 (matching the seeded event_types rows).
func TestEventTypeIDActivityCodes(t *testing.T) {
	cases := []struct {
		code string
		want int64
	}{
		{"tool_use", 4},
		{"skill", 5},
		{"blocked", 6},
		{"error", 7},
	}
	for _, c := range cases {
		if got := eventTypeID(c.code); got != c.want {
			t.Errorf("eventTypeID(%q) = %d, want %d", c.code, got, c.want)
		}
	}
}

// TestActivityCountersByWindowAndSession verifies ActivityCounters filters by
// both time window (sinceMs) and optional sessionID. It seeds 5 events across
// two sessions and two time points, then asserts:
//   - window since=100 excludes the ts=50 error row
//   - session "s1" filter excludes the s2 error row
func TestActivityCountersByWindowAndSession(t *testing.T) {
	st := tempStore(t)
	mk := func(typ, sess string, ts int64, key string) model.Event {
		return model.Event{AgentCode: "claude", TypeCode: typ, SourceCode: "transcript", SessionID: sess, TsMs: ts, DedupeKey: key}
	}
	_, err := st.ApplyIngest(IngestBatch{Events: []model.Event{
		mk("prompt", "s1", 1000, "k1"),
		mk("tool_use", "s1", 1100, "k2"),
		mk("skill", "s1", 1200, "k3"),
		mk("error", "s2", 50, "k4"), // old + different session
		mk("blocked", "s1", 1300, "k5"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	// 24h-style window: since=100 → excludes the ts=50 error.
	all, err := st.ActivityCounters(100, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if all.Prompts != 1 || all.Tools != 1 || all.Skills != 1 || all.Blocked != 1 || all.Errors != 0 {
		t.Fatalf("window counts wrong: %+v", all)
	}
	// Session s1 since=0.
	sess, _ := st.ActivityCounters(0, "s1", "")
	if sess.Errors != 0 || sess.Blocked != 1 || sess.Prompts != 1 {
		t.Fatalf("session counts wrong: %+v", sess)
	}
}

// TestSkillTriggersDeadFlag verifies that SkillTriggers returns active skills
// ordered by count desc then name, with Dead=true for skills that have never
// fired (trigger_count_total == 0 via the LEFT JOIN / COALESCE idiom).
func TestSkillTriggersDeadFlag(t *testing.T) {
	st := tempStore(t)
	// Two active skills resolved for root "" : "deploy" and "archive-old".
	items := []model.InventoryItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "deploy", Scope: model.ScopeUser, Enabled: true},
		{AgentCode: "claude", Category: model.CatSkill, Name: "archive-old", Scope: model.ScopeUser, Enabled: true},
	}
	if err := st.ReplaceInventory("claude", "", items); err != nil {
		t.Fatal(err)
	}
	if err := st.ReplaceResolved("claude", "", []model.ResolvedItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "deploy", EffectiveStatus: model.StatusActive, Winner: &items[0]},
		{AgentCode: "claude", Category: model.CatSkill, Name: "archive-old", EffectiveStatus: model.StatusActive, Winner: &items[1]},
	}); err != nil {
		t.Fatal(err)
	}
	// "deploy" fired 3×; "archive-old" never.
	if _, err := st.ApplyIngest(IngestBatch{Events: []model.Event{
		{AgentCode: "claude", TypeCode: "skill", SourceCode: "transcript", SessionID: "s1", TsMs: 5, SkillName: "deploy", DedupeKey: "a"},
		{AgentCode: "claude", TypeCode: "skill", SourceCode: "transcript", SessionID: "s1", TsMs: 6, SkillName: "deploy", DedupeKey: "b"},
		{AgentCode: "claude", TypeCode: "skill", SourceCode: "transcript", SessionID: "s1", TsMs: 7, SkillName: "deploy", DedupeKey: "c"},
	}}); err != nil {
		t.Fatal(err)
	}
	got, err := st.SkillTriggers("")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 skills, got %d (%+v)", len(got), got)
	}
	if got[0].Name != "deploy" || got[0].Count != 3 || got[0].Dead {
		t.Fatalf("row0 = %+v", got[0])
	}
	if got[1].Name != "archive-old" || got[1].Count != 0 || !got[1].Dead {
		t.Fatalf("row1 = %+v", got[1])
	}
}

// TestSkillTriggersIgnoresOtherProjectStats reproduces the latent cross-project
// contamination bug: the original JOIN ON clause omitted project_root, so a
// skill_stats row for a DIFFERENT project_root would match the same
// (agent_id, skill_name) and corrupt the returned count (and potentially
// produce duplicate rows). This test seeds a "deploy" skill for project_root=""
// and inserts a skill_stats row for project_root="other" with count=99, then
// asserts that SkillTriggers("") returns exactly ONE row with the correct count
// — NOT the 99 from the alien project.
func TestSkillTriggersIgnoresOtherProjectStats(t *testing.T) {
	st := tempStore(t)

	// Seed one active skill "deploy" for project_root "".
	items := []model.InventoryItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "deploy", Scope: model.ScopeUser, Enabled: true},
	}
	if err := st.ReplaceInventory("claude", "", items); err != nil {
		t.Fatal(err)
	}
	if err := st.ReplaceResolved("claude", "", []model.ResolvedItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "deploy", EffectiveStatus: model.StatusActive, Winner: &items[0]},
	}); err != nil {
		t.Fatal(err)
	}

	// Fire "deploy" twice via ApplyIngest so skill_stats has project_root='', count=2.
	if _, err := st.ApplyIngest(IngestBatch{Events: []model.Event{
		{AgentCode: "claude", TypeCode: "skill", SourceCode: "transcript", SessionID: "s1", TsMs: 10, SkillName: "deploy", DedupeKey: "d1"},
		{AgentCode: "claude", TypeCode: "skill", SourceCode: "transcript", SessionID: "s1", TsMs: 11, SkillName: "deploy", DedupeKey: "d2"},
	}}); err != nil {
		t.Fatal(err)
	}

	// Directly insert a skill_stats row for a DIFFERENT project_root with the
	// same agent_id=1 and skill_name="deploy". With the broken 2-column JOIN
	// (omitting project_root) this row matches the inventory_resolved row,
	// corrupting the count or producing a duplicate result row.
	if _, err := st.db.Exec(
		`INSERT INTO skill_stats (agent_id, project_root, skill_name, trigger_count_total, first_fired_ms, last_fired_ms)
		 VALUES (1, 'other', 'deploy', 99, 5, 5)`); err != nil {
		t.Fatalf("insert alien skill_stats row: %v", err)
	}

	got, err := st.SkillTriggers("")
	if err != nil {
		t.Fatal(err)
	}
	// Must return exactly one row — the alien project_root="other" row must NOT
	// appear as a second result and must NOT inflate the count.
	if len(got) != 1 {
		t.Fatalf("want 1 skill row, got %d: %+v", len(got), got)
	}
	// The count must be 2 (the project_root="" rows fired above), not 99.
	if got[0].Name != "deploy" || got[0].Count != 2 {
		t.Fatalf("want deploy count=2, got %+v", got[0])
	}
}

// TestEventRatePerMinute verifies EventRatePerMinute returns a dense, zero-filled
// slice of length `buckets`, oldest→newest, with counts only for events whose
// ts_ms falls within the last `buckets` minutes ending at nowMs. The event
// outside the window (30 minutes ago) must be excluded.
func TestEventRatePerMinute(t *testing.T) {
	st := tempStore(t)
	now := int64(10_000_000) // arbitrary ms
	minute := int64(60_000)
	if _, err := st.ApplyIngest(IngestBatch{Events: []model.Event{
		// Outside the 3-bucket window (30 minutes ago) — must be excluded.
		{AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript", SessionID: "s", TsMs: now - 30*minute, DedupeKey: "old"},
		// Two events in the "1 minute ago" bucket.
		{AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript", SessionID: "s", TsMs: now - minute, DedupeKey: "p1"},
		{AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript", SessionID: "s", TsMs: now - minute + 5, DedupeKey: "p2"},
		// One event in the "current minute" bucket.
		{AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript", SessionID: "s", TsMs: now, DedupeKey: "p3"},
	}}); err != nil {
		t.Fatal(err)
	}
	got, err := st.EventRatePerMinute("prompt", now, 3, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("want 3 buckets, got %d", len(got))
	}
	// Expected: oldest→newest: [2 min ago]=0, [1 min ago]=2, [current minute]=1
	if got[0] != 0 || got[1] != 2 || got[2] != 1 {
		t.Fatalf("buckets = %v, want [0 2 1]", got)
	}
}

// TestRecentEventsRichFields verifies that RecentEvents populates the new
// ToolName, SkillName, and Status fields from the events table columns added in
// migration 0003. A prompt row leaves all three empty (omitempty in JSON); a
// tool_use row sets ToolName; a skill row sets SkillName.
func TestRecentEventsRichFields(t *testing.T) {
	st := tempStore(t)
	if _, err := st.ApplyIngest(IngestBatch{Events: []model.Event{
		{AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript", SessionID: "s1", TsMs: 1, DedupeKey: "q1"},
		{AgentCode: "claude", TypeCode: "tool_use", SourceCode: "transcript", SessionID: "s1", TsMs: 2, ToolName: "Bash", DedupeKey: "q2"},
		{AgentCode: "claude", TypeCode: "skill", SourceCode: "transcript", SessionID: "s1", TsMs: 3, SkillName: "deploy", DedupeKey: "q3"},
	}}); err != nil {
		t.Fatal(err)
	}
	got, err := st.RecentEvents(10, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("want 3 events, got %d", len(got))
	}
	// Newest first: skill (ts=3), tool_use (ts=2), prompt (ts=1).
	if got[0].SkillName != "deploy" {
		t.Errorf("got[0].SkillName = %q, want %q", got[0].SkillName, "deploy")
	}
	if got[1].ToolName != "Bash" {
		t.Errorf("got[1].ToolName = %q, want %q", got[1].ToolName, "Bash")
	}
	if got[2].ToolName != "" || got[2].SkillName != "" || got[2].Status != "" {
		t.Errorf("prompt row should have empty rich fields: %+v", got[2])
	}
}

// TestEventsPagePaginatesByID verifies EventsPage returns a newest-first page of
// events and correctly applies both the limit and the optional type-code filter.
// Five tool_use events are seeded; EventsPage("",2,0) must return exactly 2 (the
// two newest by id); EventsPage("prompt",10,0) must return 0 because no prompt
// events exist.
func TestEventsPagePaginatesByID(t *testing.T) {
	st := tempStore(t)
	evs := make([]model.Event, 0, 5)
	for i := 0; i < 5; i++ {
		evs = append(evs, model.Event{AgentCode: "claude", TypeCode: "tool_use", SourceCode: "transcript",
			SessionID: "s", TsMs: int64(1000 + i), ToolName: "Bash", DedupeKey: string(rune('a' + i))})
	}
	if _, err := st.ApplyIngest(IngestBatch{Events: evs}); err != nil {
		t.Fatal(err)
	}
	first, err := st.EventsPage("", 2, 0, "")
	if err != nil || len(first) != 2 {
		t.Fatalf("first page len=%d err=%v", len(first), err)
	}
	if first[0].TsMs != 1004 || first[1].TsMs != 1003 {
		t.Fatalf("not newest-first: got TsMs %d, %d; want 1004, 1003", first[0].TsMs, first[1].TsMs)
	}
	// All same type filter returns only tool_use; filter mismatch returns none.
	none, err := st.EventsPage("prompt", 10, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(none) != 0 {
		t.Fatalf("type filter should exclude all, got %d", len(none))
	}
}

func TestMigration0003Schema(t *testing.T) {
	st := tempStore(t)
	// New columns exist on events.
	for _, col := range []string{"tool_name", "skill_name", "status"} {
		var cnt int
		err := st.db.QueryRow(
			`SELECT COUNT(*) FROM pragma_table_info('events') WHERE name=?`, col).Scan(&cnt)
		if err != nil || cnt != 1 {
			t.Fatalf("events.%s missing (cnt=%d err=%v)", col, cnt, err)
		}
	}
	// New event types seeded.
	var n int
	if err := st.db.QueryRow(
		`SELECT COUNT(*) FROM event_types WHERE code IN ('tool_use','skill','blocked','error')`).Scan(&n); err != nil || n != 4 {
		t.Fatalf("event_types seed: n=%d err=%v", n, err)
	}
	// skill_stats table exists.
	if _, err := st.db.Exec(
		`INSERT INTO skill_stats(agent_id,project_root,skill_name,trigger_count_total) VALUES(1,'','x',0)`); err != nil {
		t.Fatalf("skill_stats insert: %v", err)
	}
}

// TestSkillAnalyticsRowsAndDisabled verifies SkillAnalytics returns BOTH active
// and disabled skills (unlike SkillTriggers, which is active-only), joined to
// cumulative trigger counts, ordered by count desc then name asc. It carries
// effective_status and est_context_tokens through so the analytics layer can
// compute value ratio and flags.
func TestSkillAnalyticsRowsAndDisabled(t *testing.T) {
	st := tempStore(t)
	items := []model.InventoryItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "deploy", Scope: model.ScopeUser, Enabled: true},
		{AgentCode: "claude", Category: model.CatSkill, Name: "archive-old", Scope: model.ScopeUser, Enabled: true},
		{AgentCode: "claude", Category: model.CatSkill, Name: "legacy", Scope: model.ScopeUser, Enabled: false},
	}
	if err := st.ReplaceInventory("claude", "", items); err != nil {
		t.Fatal(err)
	}
	// deploy: active, 4000 tokens. archive-old: active, 1000 tokens, never fires.
	// legacy: disabled, 0 tokens (matches production: non-active items estimate 0).
	if err := st.ReplaceResolved("claude", "", []model.ResolvedItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "deploy", EffectiveStatus: model.StatusActive, Winner: &items[0], EstContextTokens: 4000},
		{AgentCode: "claude", Category: model.CatSkill, Name: "archive-old", EffectiveStatus: model.StatusActive, Winner: &items[1], EstContextTokens: 1000},
		{AgentCode: "claude", Category: model.CatSkill, Name: "legacy", EffectiveStatus: model.StatusDisabled, Winner: nil, EstContextTokens: 0},
	}); err != nil {
		t.Fatal(err)
	}
	// deploy fires 2×; archive-old + legacy never.
	if _, err := st.ApplyIngest(IngestBatch{Events: []model.Event{
		{AgentCode: "claude", TypeCode: "skill", SourceCode: "transcript", SessionID: "s1", TsMs: 5, SkillName: "deploy", DedupeKey: "a"},
		{AgentCode: "claude", TypeCode: "skill", SourceCode: "transcript", SessionID: "s1", TsMs: 6, SkillName: "deploy", DedupeKey: "b"},
	}}); err != nil {
		t.Fatal(err)
	}
	got, err := st.SkillAnalytics("")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("want 3 rows, got %d (%+v)", len(got), got)
	}
	// Order: deploy(cnt 2) first; then cnt-0 rows alphabetically: archive-old, legacy.
	if got[0].Name != "deploy" || got[0].Triggers != 2 || got[0].EstContextTokens != 4000 || got[0].EffectiveStatus != model.StatusActive {
		t.Fatalf("row0 = %+v", got[0])
	}
	if got[1].Name != "archive-old" || got[1].Triggers != 0 || got[1].EstContextTokens != 1000 {
		t.Fatalf("row1 = %+v", got[1])
	}
	if got[2].Name != "legacy" || got[2].EffectiveStatus != model.StatusDisabled || got[2].EstContextTokens != 0 {
		t.Fatalf("row2 = %+v", got[2])
	}
}

// TestSkillAnalyticsCrossProjectIsolation mirrors the SkillTriggers isolation
// test: a skill_stats row under a DIFFERENT project_root must not bleed into
// the result for project_root="" (the 3-column PK join prevents it).
func TestSkillAnalyticsCrossProjectIsolation(t *testing.T) {
	st := tempStore(t)
	items := []model.InventoryItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "deploy", Scope: model.ScopeUser, Enabled: true},
	}
	if err := st.ReplaceInventory("claude", "", items); err != nil {
		t.Fatal(err)
	}
	if err := st.ReplaceResolved("claude", "", []model.ResolvedItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "deploy", EffectiveStatus: model.StatusActive, Winner: &items[0], EstContextTokens: 500},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.ApplyIngest(IngestBatch{Events: []model.Event{
		{AgentCode: "claude", TypeCode: "skill", SourceCode: "transcript", SessionID: "s1", TsMs: 10, SkillName: "deploy", DedupeKey: "d1"},
		{AgentCode: "claude", TypeCode: "skill", SourceCode: "transcript", SessionID: "s1", TsMs: 11, SkillName: "deploy", DedupeKey: "d2"},
	}}); err != nil {
		t.Fatal(err)
	}
	// Alien row under project_root="other" with a huge count.
	if _, err := st.db.Exec(
		`INSERT INTO skill_stats (agent_id, project_root, skill_name, trigger_count_total, first_fired_ms, last_fired_ms)
		 VALUES (1, 'other', 'deploy', 99, 5, 5)`); err != nil {
		t.Fatalf("insert alien skill_stats row: %v", err)
	}
	got, err := st.SkillAnalytics("")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 row, got %d: %+v", len(got), got)
	}
	if got[0].Name != "deploy" || got[0].Triggers != 2 {
		t.Fatalf("want deploy triggers=2 (not 99), got %+v", got[0])
	}
}
