package security

import (
	"path/filepath"
	"testing"
)

// TestWriteRulesRoundTrips writes a Rules slice to a temp file, then reads it
// back with LoadRulesFromPath and asserts the IDs and fields survive the
// TOML encode/decode cycle. This guards the atomic-write path end-to-end.
func TestWriteRulesRoundTrips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "security-rules.toml")
	want := Rules{
		{
			ID:          "test-forbid-mode",
			Type:        "forbid-mode",
			Enabled:     true,
			Severity:    "high",
			Title:       "Test forbid mode",
			Remediation: "Remove it.",
			Modes:       []string{"bypassPermissions"},
		},
		{
			ID:          "test-broad-allow",
			Type:        "broad-allow",
			Enabled:     false,
			Severity:    "low",
			Title:       "Test broad allow",
			Remediation: "Narrow your allow rules.",
		},
	}
	if err := WriteRules(path, want); err != nil {
		t.Fatalf("WriteRules: %v", err)
	}
	got := LoadRulesFromPath(path, nil)
	if len(got) != 2 {
		t.Fatalf("round-trip: got %d rules, want 2", len(got))
	}
	if got[0].ID != "test-forbid-mode" || got[0].Severity != "high" {
		t.Errorf("rule[0] mismatch: %+v", got[0])
	}
	if got[1].ID != "test-broad-allow" || got[1].Enabled != false {
		t.Errorf("rule[1] mismatch: %+v", got[1])
	}
}
