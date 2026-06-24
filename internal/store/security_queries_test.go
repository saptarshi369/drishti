package store

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func TestReplaceSecurityFindings_FullReplaceAndList(t *testing.T) {
	// tempStore opens a real migrated SQLite DB in a temp dir and registers
	// cleanup. All migrations — including 0005_security.sql — are applied.
	st := tempStore(t)
	f1 := []model.Finding{
		{RuleID: "missing-env-deny", Severity: "high", Title: "x", TargetKey: "global", Detail: "d", Remediation: "r", Scope: "all"},
		{RuleID: "bypass", Severity: "high", Title: "y", TargetKey: "user:settings.json", Detail: "d2", Remediation: "r2", Scope: "user"},
	}
	if err := st.ReplaceSecurityFindings("claude", "", f1); err != nil {
		t.Fatal(err)
	}
	got, err := st.ListFindings("claude", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	// Replace with a single finding → the prior set is gone (full-replace semantics).
	if err := st.ReplaceSecurityFindings("claude", "", f1[:1]); err != nil {
		t.Fatal(err)
	}
	got, _ = st.ListFindings("claude", "")
	if len(got) != 1 || got[0].RuleID != "missing-env-deny" {
		t.Fatalf("after replace got %+v, want only missing-env-deny", got)
	}
}
