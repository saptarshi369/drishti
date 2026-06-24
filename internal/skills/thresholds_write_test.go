package skills

import (
	"path/filepath"
	"testing"
)

// TestWriteThresholdsRoundTrips writes a Thresholds value to a temp file, then
// reads it back with LoadThresholdsFromPath and asserts the values are preserved.
// This guards the atomic-write + TOML-encode path end-to-end.
func TestWriteThresholdsRoundTrips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skills-analytics.toml")
	want := Thresholds{HighTriggerMin: 25, LowValueRatioMax: 0.4}
	if err := WriteThresholds(path, want); err != nil {
		t.Fatalf("WriteThresholds: %v", err)
	}
	got := LoadThresholdsFromPath(path, nil)
	if got.HighTriggerMin != 25 || got.LowValueRatioMax != 0.4 {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}
