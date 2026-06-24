// Package settings turns raw UI input into a validated config.Config and builds
// the read-side view served at GET /api/settings. It is pure: no file or network
// I/O, so it is trivially testable. The api layer calls config.Save on the
// returned Config.
package settings

import (
	"fmt"
	"time"

	"github.com/saptarshi369/drishti/internal/config"
)

// Input is the raw, UI-supplied settings payload. Durations are strings so the
// UI can send human text like "48h"; Validate parses and clamps them. Using
// strings at the JSON boundary avoids Go's time.Duration wire encoding (which
// is nanoseconds as a number) and lets the web UI stay readable.
type Input struct {
	// Port is the TCP port the daemon listens on. 0 means "no change".
	Port int `json:"port"`

	// BindAddr is the interface address to bind (e.g. "127.0.0.1"). Empty means
	// "no change".
	BindAddr string `json:"bind_addr"`

	// Theme must be "dark" or "light". Empty means "no change".
	Theme string `json:"theme"`

	// Accent must be "default", "teal", or "violet". Empty means "no change".
	Accent string `json:"accent"`

	// ActiveWindow is a duration string (e.g. "48h") for the live-activity
	// lookback. Clamped to [1h, 168h]. Empty means "no change".
	ActiveWindow string `json:"active_window"`

	// AggregateHorizon is a duration string for how far back usage/cost stats
	// are kept. Empty means "no change".
	AggregateHorizon string `json:"aggregate_horizon"`

	// Throttle is a duration string for the daemon scheduler tick. Empty means
	// "no change".
	Throttle string `json:"throttle"`

	// ContextWindowTokens is the Claude context-window size denominator used
	// for the budget-tax %. 0 means "no change".
	ContextWindowTokens int `json:"context_window_tokens"`

	// AutoCheck opts in to a periodic GitHub releases version check. This is
	// the only outbound call Drishti ever makes.
	//
	// This field uses a pointer so the JSON decoder can distinguish "key absent"
	// (nil — keep the current value) from "key present with value false" (&false
	// — explicitly disable). A plain bool can't make that distinction: missing
	// keys decode to the zero value (false), which would silently disable
	// AutoCheck on every partial PUT that omits this key.
	AutoCheck *bool `json:"auto_check"`

	// CheckInterval is a duration string for how often to check for updates.
	// Empty means "no change".
	CheckInterval string `json:"check_interval"`
}

// Validate applies in onto a copy of cur, returning the updated Config plus any
// non-fatal Warnings. error is returned ONLY for structurally invalid input
// (unparseable duration, out-of-range port, unrecognised theme/accent).
// Clamping (e.g. active_window 999h → 168h) produces a Warning, not an error;
// the daemon can still start with a clamped value, so it is non-fatal.
//
// Validate always works on a copy of cur (out := cur), so cur is never mutated.
// This makes it safe to call speculatively without side-effects.
func Validate(cur config.Config, in Input) (config.Config, []config.Warning, error) {
	// Work on a copy so cur is never mutated if we return an error mid-way.
	out := cur
	var warns []config.Warning

	// --- port ------------------------------------------------------------------
	// Port 0 is the sentinel for "not provided"; the UI omits it or sends 0.
	if in.Port != 0 {
		if in.Port < 1 || in.Port > 65535 {
			return cur, nil, fmt.Errorf("port %d out of range [1, 65535]", in.Port)
		}
		out.Port = in.Port
	}

	// --- bind_addr -------------------------------------------------------------
	if in.BindAddr != "" {
		out.BindAddr = in.BindAddr
	}

	// --- theme -----------------------------------------------------------------
	// Theme is an enum {dark, light}; an unrecognised value is a hard error
	// because the UI is responsible for restricting choices. Warn-on-file-load
	// (config.Load) vs error-on-API (Validate) is intentional: a TOML typo by
	// a human is degraded gracefully; an API caller sending an unknown value is
	// a bug.
	if in.Theme != "" {
		switch in.Theme {
		case "dark", "light":
			out.Theme = in.Theme
		default:
			return cur, nil, fmt.Errorf("unknown theme %q; must be dark or light", in.Theme)
		}
	}

	// --- accent ----------------------------------------------------------------
	if in.Accent != "" {
		switch in.Accent {
		case "default", "teal", "violet":
			out.Accent = in.Accent
		default:
			return cur, nil, fmt.Errorf("unknown accent %q; must be default, teal, or violet", in.Accent)
		}
	}

	// --- active_window ---------------------------------------------------------
	// Parse the duration string; clamp to [1h, 168h] via config.ClampWindow,
	// which appends Warnings for any clamping it performs.
	if in.ActiveWindow != "" {
		d, err := time.ParseDuration(in.ActiveWindow)
		if err != nil {
			return cur, nil, fmt.Errorf("invalid active_window %q: %w", in.ActiveWindow, err)
		}
		clamped, cw := config.ClampWindow(d)
		out.ActiveWindow = clamped
		warns = append(warns, cw...)
	}

	// --- aggregate_horizon -----------------------------------------------------
	if in.AggregateHorizon != "" {
		d, err := time.ParseDuration(in.AggregateHorizon)
		if err != nil {
			return cur, nil, fmt.Errorf("invalid aggregate_horizon %q: %w", in.AggregateHorizon, err)
		}
		out.AggregateHorizon = d
	}

	// --- throttle --------------------------------------------------------------
	if in.Throttle != "" {
		d, err := time.ParseDuration(in.Throttle)
		if err != nil {
			return cur, nil, fmt.Errorf("invalid throttle %q: %w", in.Throttle, err)
		}
		out.ThrottleInterval = d
	}

	// --- context_window_tokens -------------------------------------------------
	if in.ContextWindowTokens != 0 {
		out.ContextWindowTokens = in.ContextWindowTokens
	}

	// --- auto_check / check_interval ------------------------------------------
	// AutoCheck is a *bool (pointer-as-optional): nil means "not provided, keep
	// current"; non-nil means the caller explicitly set it to that value. This
	// matches the "zero/empty = no change" convention used by all other fields.
	if in.AutoCheck != nil {
		out.AutoCheck = *in.AutoCheck
	}

	if in.CheckInterval != "" {
		d, err := time.ParseDuration(in.CheckInterval)
		if err != nil {
			return cur, nil, fmt.Errorf("invalid check_interval %q: %w", in.CheckInterval, err)
		}
		out.CheckInterval = d
	}

	return out, warns, nil
}

// RestartRequired reports whether moving from old to newCfg requires a daemon
// restart. Only the bound listener (port + bind address) cannot be hot-applied;
// all other settings are picked up by the ~10 s scheduler reload cycle.
func RestartRequired(old, newCfg config.Config) bool {
	return old.Port != newCfg.Port || old.BindAddr != newCfg.BindAddr
}

// View is the GET /api/settings response payload. It combines the editable
// config fields with read-only derived fields (version, privacy posture).
type View struct {
	// Editable fields mirrored from config.Config.
	Port                int      `json:"port"`
	BindAddr            string   `json:"bind_addr"`
	Theme               string   `json:"theme"`
	Accent              string   `json:"accent"`
	ActiveWindow        string   `json:"active_window"`
	ContextWindowTokens int      `json:"context_window_tokens"`
	Roots               []string `json:"roots"`
	AutoCheck           bool     `json:"auto_check"`

	// AggregateHorizon is the duration string (e.g. "72h0m0s") for how far back
	// usage/cost stats are kept. Exposed so the UI can pre-populate the field and
	// send it back unchanged in a partial PUT (read-modify-write round-trip).
	AggregateHorizon string `json:"aggregate_horizon"`

	// Throttle is the duration string for the daemon scheduler tick interval.
	// Exposed for the same read-modify-write reason as AggregateHorizon.
	Throttle string `json:"throttle"`

	// CheckInterval is the duration string for how often to poll for new
	// versions when AutoCheck is enabled. Exposed for read-modify-write.
	CheckInterval string `json:"check_interval"`

	// Read-only derived fields.

	// DBBytes is the sum of SQLite database files (*.db, *.db-wal, *.db-shm)
	// in the DataDir. Populated by the settings handler via services.DiskEstimate.
	DBBytes int64 `json:"db_bytes"`

	// BackupBytes is the total size of all files under the backup directory.
	// Populated by the settings handler via services.DiskEstimate.
	BackupBytes int64 `json:"backup_bytes"`

	// Version is the running binary version string (from main.version).
	Version string `json:"version"`

	// ScrubLocked is always true (design decision D8): Drishti scrubs secrets
	// from all stored data; this cannot be turned off by the user.
	ScrubLocked bool `json:"scrub_locked"`

	// OutboundDefaultOff is always true: no outbound connections are made
	// unless the user explicitly opts in (e.g. AutoCheck).
	OutboundDefaultOff bool `json:"outbound_default_off"`

	// MCPServers is a read-only list of configured MCP server names sourced from
	// the inventory the daemon already maintains. It is populated by the settings
	// handler after BuildView (see api/settings.go). A live `claude mcp list`
	// connection probe is deferred; this field reflects inventory records only.
	// Always a non-nil slice so JSON encodes as [] rather than null.
	MCPServers []string `json:"mcp_servers"`
}

// BuildView assembles the read-side settings view from a live config.Config and
// the running version string. Roots and MCPServers are forced to non-nil empty
// slices so JSON encodes as [] rather than null — important for UI list rendering.
func BuildView(cfg config.Config, version string) View {
	// Ensure Roots is never nil: JSON null vs [] confuses most UI frameworks.
	roots := cfg.Roots
	if roots == nil {
		roots = []string{}
	}
	return View{
		Port:                cfg.Port,
		BindAddr:            cfg.BindAddr,
		Theme:               cfg.Theme,
		Accent:              cfg.Accent,
		ActiveWindow:        cfg.ActiveWindow.String(),
		ContextWindowTokens: cfg.ContextWindowTokens,
		Roots:               roots,
		AutoCheck:           cfg.AutoCheck,
		AggregateHorizon:    cfg.AggregateHorizon.String(),
		Throttle:            cfg.ThrottleInterval.String(),
		CheckInterval:       cfg.CheckInterval.String(),
		Version:             version,
		ScrubLocked:         true,
		OutboundDefaultOff:  true,
		// MCPServers is initialised to a non-nil empty slice here so the handler
		// can append to it or leave it as-is; either way JSON encodes as [] not null.
		MCPServers: []string{},
	}
}
