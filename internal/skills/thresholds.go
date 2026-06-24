// Package skills holds the Skills Analytics engine: a user-editable thresholds
// config (loaded from TOML, with a go:embed default) and a pure function that
// derives per-skill value ratios and hygiene flags from raw store rows. It does
// no I/O beyond reading the thresholds file.
package skills

import (
	"bytes"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

//go:embed skills-analytics.toml
var defaultThresholdsTOML []byte

// Thresholds tunes the over-triggering flag (see BuildAnalytics). A skill is
// flagged when it fired at least HighTriggerMin times AND its value ratio is
// below LowValueRatioMax. Both come from the user-editable skills-analytics.toml.
// The json tags ensure the Settings API uses the same snake_case names as the
// TOML file, so the UI and the file format are consistent.
type Thresholds struct {
	HighTriggerMin   int     `toml:"high_trigger_min"   json:"high_trigger_min"`
	LowValueRatioMax float64 `toml:"low_value_ratio_max" json:"low_value_ratio_max"`
}

// Valid reports whether both knobs are positive. Zero/negative values would
// make the over-triggering rule meaningless (e.g. a ratio bar of <0 never
// fires), so such a file is rejected in favour of the embedded default and
// the Settings UI rejects such values with a 400. It is exported so the API
// handler can validate user input without duplicating the range logic.
func (t Thresholds) Valid() bool {
	return t.HighTriggerMin > 0 && t.LowValueRatioMax > 0
}

// DefaultThresholds returns the embedded default thresholds. The embedded file
// is validated by a unit test, so decoding is not expected to fail; if it
// somehow does, the zero value is returned (a safe degrade: nothing is flagged
// over-triggering, rather than a crash — spec §14).
func DefaultThresholds() Thresholds {
	var t Thresholds
	if _, err := toml.Decode(string(defaultThresholdsTOML), &t); err != nil {
		return Thresholds{}
	}
	return t
}

// LoadThresholdsFromPath reads the user-editable thresholds file at path. On any
// problem — missing, unreadable, malformed, or out-of-range values — it logs a
// warning (when lg is non-nil) and falls back to the embedded default so the
// Skills screen never breaks on a bad file (spec §14).
func LoadThresholdsFromPath(path string, lg *slog.Logger) Thresholds {
	data, err := os.ReadFile(path)
	if err != nil {
		if lg != nil {
			lg.Warn("skills thresholds unreadable; using built-in defaults", "path", path, "err", err)
		}
		return DefaultThresholds()
	}
	var t Thresholds
	if _, err := toml.Decode(string(data), &t); err != nil {
		if lg != nil {
			lg.Warn("skills thresholds malformed; using built-in defaults", "path", path, "err", err)
		}
		return DefaultThresholds()
	}
	if !t.Valid() {
		if lg != nil {
			lg.Warn("skills thresholds out of range; using built-in defaults", "path", path, "value", t)
		}
		return DefaultThresholds()
	}
	return t
}

// EnsureThresholdsFile writes the embedded default to path when no file exists
// there, giving the user a documented file to edit. An existing file is left
// untouched so user edits are never clobbered. The parent directory is created
// (mode 0755) so callers need not pre-create it.
func EnsureThresholdsFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, defaultThresholdsTOML, 0o644)
}

// WriteThresholds writes t to path as TOML, prefixed with a generated header
// comment that identifies the file and explains that changes are picked up
// automatically within ~10 s by the Drishti daemon scheduler. The write is
// atomic: the encoder first fills an in-memory buffer, then the bytes are
// written to a temp file in the same directory and renamed over the target —
// a reader will always see either the old or the new file, never a partial one
// (same pattern as config.Save; see internal/config/config.go).
func WriteThresholds(path string, t Thresholds) error {
	// Header comment: documents the file and mentions the hot-reload cadence so
	// a user who opens the file manually knows it is managed by Drishti Settings.
	const header = "# skills-analytics.toml — managed by Drishti.\n" +
		"# Edit here or via the Settings UI; changes apply automatically within ~10s.\n\n"

	var buf bytes.Buffer
	buf.WriteString(header)
	// toml.NewEncoder encodes the struct using the toml struct tags on Thresholds
	// (high_trigger_min, low_value_ratio_max). Encode returns an error only if
	// the value cannot be represented as TOML (not possible for a simple struct).
	if err := toml.NewEncoder(&buf).Encode(t); err != nil {
		return fmt.Errorf("encode thresholds: %w", err)
	}
	return atomicWriteFile(path, buf.Bytes())
}

// atomicWriteFile writes data to path via a temp file + os.Rename. Using the
// same directory as path ensures src and dst are on the same filesystem, which
// is required for os.Rename to be atomic (POSIX guarantee). A crash after
// write but before rename leaves a stale .tmp file but never a half-written
// target — the original remains intact until the rename succeeds.
func atomicWriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	// Create the temp file in the same dir as the target (same mount point).
	tmp, err := os.CreateTemp(dir, ".write-*.tmp")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	tmpName := tmp.Name()
	// Clean up on any error after this point.
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()        //nolint:errcheck // already in error path
		os.Remove(tmpName) //nolint:errcheck // best-effort cleanup
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName) //nolint:errcheck // best-effort cleanup
		return fmt.Errorf("close temp: %w", err)
	}
	// Atomic rename: the kernel commits this in one syscall, so readers always
	// see a complete file (either old or new).
	return os.Rename(tmpName, path)
}
