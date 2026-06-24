package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	dir := t.TempDir()
	cfg, warns, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 7777 {
		t.Errorf("port = %d, want 7777", cfg.Port)
	}
	if cfg.BindAddr != "127.0.0.1" {
		t.Errorf("bind = %q, want 127.0.0.1", cfg.BindAddr)
	}
	if len(warns) == 0 {
		t.Errorf("expected a warning about the missing config file")
	}
}

func TestLoadClampsActiveWindow(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.toml"),
		[]byte("[retention]\nactive_window = \"720h\"\n"), 0o644)
	cfg, _, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ActiveWindow != 168*time.Hour {
		t.Errorf("active_window = %v, want clamped to 168h", cfg.ActiveWindow)
	}
}

func TestLoadInvalidTomlDegrades(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.toml"), []byte("this is not toml ::: ["), 0o644)
	cfg, warns, err := Load(dir)
	if err != nil {
		t.Fatalf("invalid toml must not error, got %v", err)
	}
	if cfg.Port != 7777 || len(warns) == 0 {
		t.Errorf("invalid toml should fall back to defaults + warning")
	}
}

func TestLoad_Roots(t *testing.T) {
	dir := t.TempDir()
	body := "[roots]\npaths = [\"/tmp/projA\", \"/tmp/projB\"]\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, _, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Roots) != 2 || cfg.Roots[0] != "/tmp/projA" {
		t.Fatalf("roots = %v", cfg.Roots)
	}
}

func TestLoad_ContextWindowDefault(t *testing.T) {
	cfg, _, err := Load(t.TempDir()) // no config.toml present
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ContextWindowTokens != 200000 {
		t.Errorf("default window = %d, want 200000", cfg.ContextWindowTokens)
	}
}

func TestLoad_ContextWindowOverride(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.toml"),
		[]byte("[context]\nwindow_tokens = 1000000\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, _, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ContextWindowTokens != 1000000 {
		t.Errorf("window = %d, want 1000000", cfg.ContextWindowTokens)
	}
}

func TestLoad_ContextWindowAbsurdKeepsDefault(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.toml"),
		[]byte("[context]\nwindow_tokens = -5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, warns, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ContextWindowTokens != 200000 {
		t.Errorf("window = %d, want 200000 (default kept)", cfg.ContextWindowTokens)
	}
	if len(warns) == 0 {
		t.Error("expected a warning for the absurd window value")
	}
}

// TestLoad_ThemeValidation checks that an invalid theme value keeps the default
// ("dark") and records a warning, while a valid value is accepted.
func TestLoad_ThemeValidation(t *testing.T) {
	t.Run("valid theme accepted", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "config.toml"),
			[]byte("[ui]\ntheme = \"light\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, warns, err := Load(dir)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Theme != "light" {
			t.Errorf("theme = %q, want \"light\"", cfg.Theme)
		}
		if len(warns) != 0 {
			t.Errorf("unexpected warns: %v", warns)
		}
	})

	t.Run("invalid theme keeps default and warns", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "config.toml"),
			[]byte("[ui]\ntheme = \"neon\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, warns, err := Load(dir)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Theme != "dark" {
			t.Errorf("theme = %q, want \"dark\" (default)", cfg.Theme)
		}
		if len(warns) == 0 {
			t.Error("expected a warning for invalid theme")
		}
	})
}

// TestLoad_AccentValidation checks that an invalid accent value keeps the default
// ("default") and records a warning, while a valid value is accepted.
func TestLoad_AccentValidation(t *testing.T) {
	t.Run("valid accent accepted", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "config.toml"),
			[]byte("[ui]\naccent = \"violet\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, warns, err := Load(dir)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Accent != "violet" {
			t.Errorf("accent = %q, want \"violet\"", cfg.Accent)
		}
		if len(warns) != 0 {
			t.Errorf("unexpected warns: %v", warns)
		}
	})

	t.Run("invalid accent keeps default and warns", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "config.toml"),
			[]byte("[ui]\naccent = \"rainbow\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, warns, err := Load(dir)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Accent != "default" {
			t.Errorf("accent = %q, want \"default\"", cfg.Accent)
		}
		if len(warns) == 0 {
			t.Error("expected a warning for invalid accent")
		}
	})
}

// TestClampWindow verifies the exported ClampWindow helper enforces [1h, 168h].
func TestClampWindow(t *testing.T) {
	tests := []struct {
		name       string
		input      time.Duration
		wantOut    time.Duration
		wantWarned bool
	}{
		{"below min clamped to 1h", 30 * time.Minute, time.Hour, true},
		{"above max clamped to 168h", 999 * time.Hour, 168 * time.Hour, true},
		{"within range unchanged", 48 * time.Hour, 48 * time.Hour, false},
		{"at min boundary", time.Hour, time.Hour, false},
		{"at max boundary", 168 * time.Hour, 168 * time.Hour, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, warns := ClampWindow(tc.input)
			if got != tc.wantOut {
				t.Errorf("ClampWindow(%v) = %v, want %v", tc.input, got, tc.wantOut)
			}
			if (len(warns) > 0) != tc.wantWarned {
				t.Errorf("ClampWindow(%v) warns=%v, wantWarned=%v", tc.input, warns, tc.wantWarned)
			}
		})
	}
}
