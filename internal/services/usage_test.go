package services

import (
	"path/filepath"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/store"
)

func openStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "drishti.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

// TestQuotaSnapshotGatedWhenEmpty verifies an empty quota_samples table yields a
// snapshot with Available=false and nil windows (the UI's gated state).
func TestQuotaSnapshotGatedWhenEmpty(t *testing.T) {
	st := openStore(t)
	snap, err := QuotaSnapshot(st, "claude")
	if err != nil {
		t.Fatal(err)
	}
	if snap.Available || snap.FiveHour != nil || snap.SevenDay != nil {
		t.Fatalf("want gated snapshot, got %+v", snap)
	}
}

// TestQuotaSnapshotMapsWindows verifies samples map into the five_hour/seven_day
// pointers with Available=true and plan/source carried through.
func TestQuotaSnapshotMapsWindows(t *testing.T) {
	st := openStore(t)
	for _, r := range []model.QuotaSampleRow{
		{AgentCode: "claude", Window: "five_hour", UsedPercentage: 41, ResetsAtMs: 9, TsMs: 2000, Plan: "max", Source: "statusline"},
		{AgentCode: "claude", Window: "seven_day", UsedPercentage: 12, ResetsAtMs: 7, TsMs: 1500, Plan: "max", Source: "statusline"},
	} {
		if err := st.InsertQuotaSample(r); err != nil {
			t.Fatal(err)
		}
	}
	snap, err := QuotaSnapshot(st, "claude")
	if err != nil {
		t.Fatal(err)
	}
	if !snap.Available || snap.FiveHour == nil || snap.SevenDay == nil {
		t.Fatalf("want both windows available, got %+v", snap)
	}
	if snap.FiveHour.UsedPercentage != 41 || snap.Plan != "max" {
		t.Fatalf("five_hour/plan wrong: %+v", snap)
	}
}

// TestUsageSnapshotShape verifies the 7-day zero-fill, totals, and by-project /
// by-model percentages from seeded rollup rows. Days are seeded relative to today
// so the test is stable regardless of when it runs.
func TestUsageSnapshotShape(t *testing.T) {
	st := openStore(t)
	today := dayToTime(todayDay())
	dayInt := func(daysAgo int) int {
		d := today.AddDate(0, 0, -daysAgo)
		return d.Year()*10000 + int(d.Month())*100 + d.Day()
	}
	seed := func(day int, root, mdl string, in, out, cache int64) {
		total := in + out + cache
		if _, err := st.DB().Exec(
			`INSERT INTO usage_rollup (agent_id, day, project_root, model,
			   input_tokens, output_tokens, cache_tokens, total_tokens)
			 VALUES (1,?,?,?,?,?,?,?)`, day, root, mdl, in, out, cache, total); err != nil {
			t.Fatal(err)
		}
	}
	seed(dayInt(0), "-myapp", "claude-opus-4-8", 100, 20, 0)
	seed(dayInt(2), "-website", "claude-sonnet-4-6", 50, 5, 0)

	snap, err := UsageSnapshot(st, "claude")
	if err != nil {
		t.Fatal(err)
	}
	if snap.WindowDays != 7 || len(snap.Days) != 7 {
		t.Fatalf("want 7 zero-filled days, got %d", len(snap.Days))
	}
	if !snap.Estimate {
		t.Fatal("Estimate must be true")
	}
	if snap.TotalTokens != 100+20+50+5 {
		t.Fatalf("total tokens = %d", snap.TotalTokens)
	}
	if len(snap.ByProject) != 2 {
		t.Fatalf("byProject = %+v", snap.ByProject)
	}
	// Labels are the last '-' segment of the encoded root.
	if snap.ByProject[0].Name != "myapp" && snap.ByProject[1].Name != "myapp" {
		t.Fatalf("expected a 'myapp' label, got %+v", snap.ByProject)
	}
	if len(snap.Heatmap) != 56 {
		t.Fatalf("want 56-day heatmap, got %d", len(snap.Heatmap))
	}
	if len(snap.ByModel) != 2 {
		t.Fatalf("want 2 by-model entries, got %+v", snap.ByModel)
	}
}

// TestUsageSnapshotMergesModelsByLabel verifies that two DIFFERENT raw model ids
// which map to the same display label (e.g. "claude-opus-4-8" and an older
// "claude-3-opus-…" both → "Opus") are merged into ONE by-model row. Without the
// merge the breakdown showed two "Opus" rows, and the UI's name-keyed {#each}
// crashed with each_key_duplicate — leaving the Usage page stuck on "Loading…".
func TestUsageSnapshotMergesModelsByLabel(t *testing.T) {
	st := openStore(t)
	today := dayToTime(todayDay())
	todayInt := today.Year()*10000 + int(today.Month())*100 + today.Day()
	seed := func(mdl string, in int64) {
		if _, err := st.DB().Exec(
			`INSERT INTO usage_rollup (agent_id, day, project_root, model,
			   input_tokens, output_tokens, cache_tokens, total_tokens)
			 VALUES (1,?,?,?,?,0,0,?)`, todayInt, "-proj", mdl, in, in); err != nil {
			t.Fatal(err)
		}
	}
	// Two distinct Opus builds + one Sonnet.
	seed("claude-opus-4-8", 60)
	seed("claude-3-opus-20240229", 20)
	seed("claude-sonnet-4-6", 20)

	snap, err := UsageSnapshot(st, "claude")
	if err != nil {
		t.Fatal(err)
	}
	// Names must be unique (no duplicate key for the UI's name-keyed list).
	seen := map[string]bool{}
	opus := 0
	for _, m := range snap.ByModel {
		if seen[m.Name] {
			t.Errorf("duplicate by-model label %q in %+v", m.Name, snap.ByModel)
		}
		seen[m.Name] = true
		if m.Name == "Opus" {
			opus++
		}
	}
	if opus != 1 {
		t.Errorf("want exactly one merged Opus row, got %d in %+v", opus, snap.ByModel)
	}
	// Merged Opus = 80 of 100 tokens → 80%.
	for _, m := range snap.ByModel {
		if m.Name == "Opus" && m.Pct != 80 {
			t.Errorf("merged Opus pct = %d, want 80", m.Pct)
		}
	}
}

// TestUsageSnapshotStreak verifies that StreakDays is correctly computed by
// UsageSnapshot when today has data but yesterday does not.
func TestUsageSnapshotStreak(t *testing.T) {
	st := openStore(t)
	today := dayToTime(todayDay())
	todayInt := today.Year()*10000 + int(today.Month())*100 + today.Day()
	if _, err := st.DB().Exec(
		`INSERT INTO usage_rollup (agent_id, day, project_root, model,
		   input_tokens, output_tokens, cache_tokens, total_tokens)
		 VALUES (1,?,?,?,?,?,?,?)`, todayInt, "-proj", "claude-opus-4-8", 10, 2, 0, 12); err != nil {
		t.Fatal(err)
	}
	snap, err := UsageSnapshot(st, "claude")
	if err != nil {
		t.Fatal(err)
	}
	if snap.StreakDays != 1 {
		t.Fatalf("want streak 1 (today active, yesterday not), got %d", snap.StreakDays)
	}
}

// TestHeatBucket verifies intensity bucketing relative to the window max.
func TestHeatBucket(t *testing.T) {
	cases := []struct {
		total, max int64
		want       int
	}{
		{0, 100, 0},  // no activity
		{20, 100, 1}, // <= 33%
		{50, 100, 2}, // <= 66%
		{90, 100, 3}, // > 66%
		{10, 0, 0},   // max 0 (no activity anywhere) → 0, never divide by zero
	}
	for _, c := range cases {
		if got := heatBucket(c.total, c.max); got != c.want {
			t.Errorf("heatBucket(%d,%d) = %d, want %d", c.total, c.max, got, c.want)
		}
	}
}

// TestComputeStreak covers the streak edge cases over a set of "active" days
// (yyyymmdd ints with total_tokens > 0), anchored to a fixed "today".
func TestComputeStreak(t *testing.T) {
	today := 20260622
	cases := []struct {
		name   string
		active map[int]bool
		want   int
	}{
		{"no data", map[int]bool{}, 0},
		{"today only", map[int]bool{20260622: true}, 1},
		{"unbroken 3", map[int]bool{20260622: true, 20260621: true, 20260620: true}, 3},
		{"gap at today counts from yesterday", map[int]bool{20260621: true, 20260620: true}, 2},
		{"broken run", map[int]bool{20260622: true, 20260620: true}, 1},
	}
	for _, c := range cases {
		if got := computeStreak(today, c.active); got != c.want {
			t.Errorf("%s: computeStreak = %d, want %d", c.name, got, c.want)
		}
	}
}
