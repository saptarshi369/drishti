// Package services assembles read-models from raw store rows for the API layer.
// Each function in this file is pure (no I/O) so it can be unit-tested without
// a real database.
package services

import "github.com/saptarshi369/drishti/internal/model"

// BuildSecuritySnapshot folds the stored findings into the read-model the
// Security screen consumes: the findings (already ordered by the store) plus
// per-severity counts and a total. Both Findings and Counts are always non-nil
// so the JSON shape is stable even for an all-clear result ("findings": [] not
// null, "counts": {} not null).
func BuildSecuritySnapshot(findings []model.Finding) model.SecuritySnapshot {
	// Guarantee a non-nil slice so json.Marshal emits [] instead of null.
	// This matters for the all-clear state where ListFindings returns nil.
	if findings == nil {
		findings = []model.Finding{}
	}

	// Walk the findings once, tallying per-severity counts into a fresh map.
	// Starting from map[string]int{} (not nil) ensures the counts field is also
	// always non-nil in the JSON output.
	counts := map[string]int{}
	for _, f := range findings {
		counts[f.Severity]++
	}

	return model.SecuritySnapshot{
		Findings: findings,
		Counts:   counts,
		Total:    len(findings),
	}
}
