package config

import (
	"testing"
	"time"
)

// TestSaveLoadRoundTrip verifies that Save writes a file Load reads back
// identically for the fields the UI can edit.
func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	in := Default()
	in.DataDir = dir
	in.Port = 8080
	in.BindAddr = "127.0.0.1"
	in.Theme = "light"
	in.Accent = "teal"
	in.ActiveWindow = 48 * time.Hour
	in.ContextWindowTokens = 150000
	in.Roots = []string{"/tmp/a", "/tmp/b"}
	in.AutoCheck = true

	if err := Save(in); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, warns, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(warns) != 0 {
		t.Fatalf("round-trip should produce no warnings, got %v", warns)
	}
	if got.Port != 8080 || got.Theme != "light" || got.Accent != "teal" {
		t.Errorf("scalar mismatch: %+v", got)
	}
	if got.ActiveWindow != 48*time.Hour || got.ContextWindowTokens != 150000 {
		t.Errorf("duration/int mismatch: %+v", got)
	}
	if len(got.Roots) != 2 || !got.AutoCheck {
		t.Errorf("roots/bool mismatch: %+v", got)
	}
}
