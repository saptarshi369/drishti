package settings

import (
	"testing"
	"time"

	"github.com/saptarshi369/drishti/internal/config"
)

func TestValidateClampsActiveWindow(t *testing.T) {
	cur := config.Default()
	out, warns, err := Validate(cur, Input{ActiveWindow: "999h"})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if out.ActiveWindow != 168*time.Hour {
		t.Errorf("active_window = %v, want clamped 168h", out.ActiveWindow)
	}
	if len(warns) == 0 {
		t.Errorf("expected a clamp warning")
	}
}

func TestRestartRequiredOnPortChange(t *testing.T) {
	a := config.Default()
	b := a
	b.Port = 9000
	if !RestartRequired(a, b) {
		t.Error("port change should require restart")
	}
	c := a
	c.Theme = "light"
	if RestartRequired(a, c) {
		t.Error("theme change should NOT require restart")
	}
}

func TestValidateInvalidPortErrors(t *testing.T) {
	cur := config.Default()
	_, _, err := Validate(cur, Input{Port: 99999})
	if err == nil {
		t.Error("port 99999 should error")
	}
}

func TestValidateInvalidThemeErrors(t *testing.T) {
	cur := config.Default()
	_, _, err := Validate(cur, Input{Theme: "neon"})
	if err == nil {
		t.Error("unknown theme should error")
	}
}

func TestValidateInvalidAccentErrors(t *testing.T) {
	cur := config.Default()
	_, _, err := Validate(cur, Input{Accent: "rainbow"})
	if err == nil {
		t.Error("unknown accent should error")
	}
}

func TestValidateUnparsableDurationErrors(t *testing.T) {
	cur := config.Default()
	_, _, err := Validate(cur, Input{ActiveWindow: "not-a-duration"})
	if err == nil {
		t.Error("bad active_window duration should error")
	}
}

func TestRestartRequiredOnBindAddrChange(t *testing.T) {
	a := config.Default()
	b := a
	b.BindAddr = "0.0.0.0"
	if !RestartRequired(a, b) {
		t.Error("bind_addr change should require restart")
	}
}

// boolPtr is a tiny test helper that returns a pointer to the given bool.
// Go does not allow taking the address of a literal (e.g. &false is illegal),
// so helpers like this are idiomatic in test files.
func boolPtr(b bool) *bool { return &b }

// TestAutoCheckNilKeepsCurrentValue verifies Fix 1: when Input.AutoCheck is nil
// (the key is absent from the JSON payload), Validate must NOT overwrite the
// current setting. Previously AutoCheck was a plain bool, so a partial PUT that
// omitted auto_check silently reset it to false.
func TestAutoCheckNilKeepsCurrentValue(t *testing.T) {
	cur := config.Default()
	cur.AutoCheck = true // start enabled

	// Input has AutoCheck == nil (not provided).
	out, _, err := Validate(cur, Input{})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !out.AutoCheck {
		t.Error("AutoCheck should remain true when Input.AutoCheck is nil (omitted)")
	}
}

// TestAutoCheckNonNilFalseDisables verifies that a non-nil *bool false in Input
// does disable AutoCheck, distinguishing explicit-false from omitted.
func TestAutoCheckNonNilFalseDisables(t *testing.T) {
	cur := config.Default()
	cur.AutoCheck = true // start enabled

	out, _, err := Validate(cur, Input{AutoCheck: boolPtr(false)})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if out.AutoCheck {
		t.Error("AutoCheck should be false when Input.AutoCheck is &false")
	}
}

// TestBuildViewExposesHorizons verifies Fix 2: View must include AggregateHorizon,
// Throttle, and CheckInterval so the UI can pre-populate those fields for a
// read-modify-write round-trip.
func TestBuildViewExposesHorizons(t *testing.T) {
	cfg := config.Default()
	cfg.AggregateHorizon = 72 * time.Hour
	cfg.ThrottleInterval = 15 * time.Second
	cfg.CheckInterval = 24 * time.Hour

	v := BuildView(cfg, "v0")

	if v.AggregateHorizon != (72 * time.Hour).String() {
		t.Errorf("AggregateHorizon = %q, want %q", v.AggregateHorizon, (72 * time.Hour).String())
	}
	if v.Throttle != (15 * time.Second).String() {
		t.Errorf("Throttle = %q, want %q", v.Throttle, (15 * time.Second).String())
	}
	if v.CheckInterval != (24 * time.Hour).String() {
		t.Errorf("CheckInterval = %q, want %q", v.CheckInterval, (24 * time.Hour).String())
	}
}

// TestRestartRequiredOnBindAddrChangeFalseArm verifies symmetry with
// TestRestartRequiredOnPortChange: a non-listener change (Theme) must return false.
func TestRestartRequiredOnBindAddrChangeFalseArm(t *testing.T) {
	a := config.Default()
	b := a
	b.Theme = "light"
	if RestartRequired(a, b) {
		t.Error("theme change should NOT require restart")
	}
}

func TestBuildViewRootsNonNil(t *testing.T) {
	cfg := config.Default()
	// cfg.Roots is nil from Default(); BuildView must return [] not null.
	v := BuildView(cfg, "v1.2.3")
	if v.Roots == nil {
		t.Error("Roots must be non-nil (JSON [] not null)")
	}
	if v.Version != "v1.2.3" {
		t.Errorf("Version = %q, want v1.2.3", v.Version)
	}
	if !v.ScrubLocked {
		t.Error("ScrubLocked must always be true")
	}
	if !v.OutboundDefaultOff {
		t.Error("OutboundDefaultOff must always be true")
	}
}
