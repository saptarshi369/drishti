package services

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

// TestBuildSecuritySnapshot_EmptyIsAllClear verifies that a nil input (no stored
// findings) produces an all-clear snapshot: Total 0, non-nil Findings slice,
// and non-nil Counts map. The non-nil guarantee keeps the JSON shape stable
// ("findings": [] rather than null).
func TestBuildSecuritySnapshot_EmptyIsAllClear(t *testing.T) {
	snap := BuildSecuritySnapshot(nil)
	if snap.Total != 0 || len(snap.Findings) != 0 {
		t.Fatalf("empty snapshot not all-clear: %+v", snap)
	}
	// Both fields must be non-nil for stable JSON serialisation.
	if snap.Counts == nil {
		t.Fatal("Counts should be a non-nil empty map for stable JSON")
	}
	if snap.Findings == nil {
		t.Fatal("Findings should be a non-nil empty slice for stable JSON")
	}
}

// TestBuildSecuritySnapshot_Counts verifies that BuildSecuritySnapshot correctly
// tallies per-severity counts and sets Total.
func TestBuildSecuritySnapshot_Counts(t *testing.T) {
	f := []model.Finding{
		{RuleID: "a", Severity: "high"},
		{RuleID: "b", Severity: "high"},
		{RuleID: "c", Severity: "low"},
	}
	snap := BuildSecuritySnapshot(f)
	if snap.Total != 3 {
		t.Fatalf("Total = %d, want 3", snap.Total)
	}
	if snap.Counts["high"] != 2 || snap.Counts["low"] != 1 {
		t.Fatalf("Counts = %v", snap.Counts)
	}
	if len(snap.Findings) != 3 {
		t.Fatalf("Findings len = %d", len(snap.Findings))
	}
}
