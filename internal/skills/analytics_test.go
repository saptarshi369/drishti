package skills

import (
	"math"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

// thr is the standard threshold set used across these tests (the documented
// defaults): over-triggering needs >=20 triggers AND value ratio < 5.0.
var thr = Thresholds{HighTriggerMin: 20, LowValueRatioMax: 5.0}

// TestBuildAnalytics_ValueRatioAndFlags walks the full flag truth table.
func TestBuildAnalytics_ValueRatioAndFlags(t *testing.T) {
	rows := []model.SkillStatRow{
		// heavy-noisy: active, 10000 tok, 30 fires → ratio 3.0 (<5) & >=20 → OVER.
		{Name: "heavy-noisy", EffectiveStatus: model.StatusActive, EstContextTokens: 10000, Triggers: 30},
		// lean-useful: active, 500 tok, 30 fires → ratio 60 → not over.
		{Name: "lean-useful", EffectiveStatus: model.StatusActive, EstContextTokens: 500, Triggers: 30},
		// dead-weight: active, 2000 tok, 0 fires → DEAD; not over (triggers < 20).
		{Name: "dead-weight", EffectiveStatus: model.StatusActive, EstContextTokens: 2000, Triggers: 0},
		// off: disabled, 0 tok, 5 historical fires → DISABLED only.
		{Name: "off", EffectiveStatus: model.StatusDisabled, EstContextTokens: 0, Triggers: 5},
	}
	snap := BuildAnalytics(rows, thr)
	by := map[string]model.SkillAnalyticsItem{}
	for _, it := range snap.Items {
		by[it.Name] = it
	}

	if got := by["heavy-noisy"]; !approx(got.ValueRatio, 3.0) || !got.OverTriggering || got.Dead || got.Disabled {
		t.Fatalf("heavy-noisy = %+v", got)
	}
	if got := by["lean-useful"]; !approx(got.ValueRatio, 60.0) || got.OverTriggering {
		t.Fatalf("lean-useful = %+v", got)
	}
	if got := by["dead-weight"]; !got.Dead || got.OverTriggering || got.ValueRatio != 0 {
		t.Fatalf("dead-weight = %+v", got)
	}
	if got := by["off"]; !got.Disabled || got.Dead || got.OverTriggering || got.ValueRatio != 0 {
		t.Fatalf("off = %+v", got)
	}
}

// TestBuildAnalytics_OverTriggeringBoundaries pins the exact threshold edges:
// triggers == HighTriggerMin is "high" (>=), ratio == LowValueRatioMax is NOT
// low (strict <).
func TestBuildAnalytics_OverTriggeringBoundaries(t *testing.T) {
	rows := []model.SkillStatRow{
		// 20 fires, 5000 tok → ratio 4.0 (<5) → OVER (triggers exactly at floor).
		{Name: "edge-in", EffectiveStatus: model.StatusActive, EstContextTokens: 5000, Triggers: 20},
		// 20 fires, 4000 tok → ratio exactly 5.0 → NOT over (strict <).
		{Name: "edge-out", EffectiveStatus: model.StatusActive, EstContextTokens: 4000, Triggers: 20},
	}
	snap := BuildAnalytics(rows, thr)
	by := map[string]model.SkillAnalyticsItem{}
	for _, it := range snap.Items {
		by[it.Name] = it
	}
	if !by["edge-in"].OverTriggering {
		t.Errorf("edge-in should be over-triggering: %+v", by["edge-in"])
	}
	if by["edge-out"].OverTriggering {
		t.Errorf("edge-out should NOT be over-triggering (ratio==max): %+v", by["edge-out"])
	}
}

// TestBuildAnalytics_AggregatesAndEmptyShape verifies the snapshot totals and
// the non-nil empty shape (stable JSON []).
func TestBuildAnalytics_AggregatesAndEmptyShape(t *testing.T) {
	empty := BuildAnalytics(nil, thr)
	if empty.Items == nil {
		t.Fatal("Items must be non-nil for stable JSON ([] not null)")
	}
	if empty.Total != 0 || empty.TotalContextTokens != 0 {
		t.Fatalf("empty totals = %+v", empty)
	}

	rows := []model.SkillStatRow{
		{Name: "a", EffectiveStatus: model.StatusActive, EstContextTokens: 1000, Triggers: 0},  // dead
		{Name: "b", EffectiveStatus: model.StatusActive, EstContextTokens: 8000, Triggers: 25}, // over (ratio 3.125)
		{Name: "c", EffectiveStatus: model.StatusDisabled, EstContextTokens: 0, Triggers: 0},   // disabled
	}
	snap := BuildAnalytics(rows, thr)
	if snap.Total != 3 || snap.TotalContextTokens != 9000 {
		t.Fatalf("totals = %+v", snap)
	}
	if snap.Counts.Dead != 1 || snap.Counts.OverTriggering != 1 || snap.Counts.Disabled != 1 {
		t.Fatalf("counts = %+v", snap.Counts)
	}
}

// approx compares floats with a small tolerance.
func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
