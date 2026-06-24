package model

import "testing"

func TestValidSeverity(t *testing.T) {
	for _, s := range []string{"high", "medium", "low"} {
		if !ValidSeverity(s) {
			t.Fatalf("ValidSeverity(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"", "critical", "warning", "HIGH"} {
		if ValidSeverity(s) {
			t.Fatalf("ValidSeverity(%q) = true, want false", s)
		}
	}
}
