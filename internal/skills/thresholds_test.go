package skills

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDefaultThresholds asserts the embedded TOML decodes to the documented,
// valid defaults. This guards the compiled-in file: if an edit breaks it, CI
// fails here rather than silently disabling the over-triggering flag.
func TestDefaultThresholds(t *testing.T) {
	d := DefaultThresholds()
	if d.HighTriggerMin != 20 || d.LowValueRatioMax != 5.0 {
		t.Fatalf("default thresholds = %+v, want {20 5}", d)
	}
}

// TestLoadThresholdsFromPath_Good reads a valid user file and returns its values.
func TestLoadThresholdsFromPath_Good(t *testing.T) {
	p := filepath.Join(t.TempDir(), "skills-analytics.toml")
	if err := os.WriteFile(p, []byte("high_trigger_min = 5\nlow_value_ratio_max = 2.5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := LoadThresholdsFromPath(p, nil)
	if got.HighTriggerMin != 5 || got.LowValueRatioMax != 2.5 {
		t.Fatalf("loaded = %+v, want {5 2.5}", got)
	}
}

// TestLoadThresholdsFromPath_FallsBack covers every failure mode — missing file,
// unparseable TOML, and out-of-range values — all of which must degrade to the
// embedded default so analytics never break on a bad config file (spec §14).
func TestLoadThresholdsFromPath_FallsBack(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]string{
		"missing":     filepath.Join(dir, "does-not-exist.toml"),
		"unparseable": writeTmp(t, dir, "bad.toml", "high_trigger_min = = ="),
		"zero":        writeTmp(t, dir, "zero.toml", "high_trigger_min = 0\nlow_value_ratio_max = 5.0\n"),
		"negative":    writeTmp(t, dir, "neg.toml", "high_trigger_min = 20\nlow_value_ratio_max = -1\n"),
	}
	def := DefaultThresholds()
	for name, path := range cases {
		got := LoadThresholdsFromPath(path, nil)
		if got != def {
			t.Errorf("%s: got %+v, want default %+v", name, got, def)
		}
	}
}

// TestEnsureThresholdsFile writes the default on absence and never clobbers an
// existing (possibly user-edited) file.
func TestEnsureThresholdsFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "sub", "skills-analytics.toml")
	if err := EnsureThresholdsFile(p); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("file not written: %v", err)
	}
	// Overwrite with a sentinel, ensure a second call leaves it untouched.
	if err := os.WriteFile(p, []byte("high_trigger_min = 7\nlow_value_ratio_max = 1.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureThresholdsFile(p); err != nil {
		t.Fatal(err)
	}
	if got := LoadThresholdsFromPath(p, nil); got.HighTriggerMin != 7 {
		t.Fatalf("EnsureThresholdsFile clobbered user edits: %+v", got)
	}
}

// writeTmp writes content to dir/name and returns the path (test helper).
func writeTmp(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}
