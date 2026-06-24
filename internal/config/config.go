// Package config loads Drishti's optional TOML configuration.
//
// The config file is OPTIONAL: if it is missing or partly invalid, Load returns
// safe defaults plus non-fatal Warnings and never an error (except for truly
// unrecoverable IO). The daemon must always be able to start.
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// Warning is a non-fatal configuration problem surfaced to the UI/logs.
type Warning string

// Config holds the full set of settings Drishti uses at runtime.
// All fields have safe defaults (see Default); the TOML file may override
// a subset of them — unmentioned fields simply keep their defaults.
type Config struct {
	// Core network settings.
	Port     int
	BindAddr string
	DataDir  string
	LogLevel string

	// UI appearance (served to the web shell; hot-applies without restart).
	// Theme is one of "dark" or "light"; Accent is one of "default", "teal", "violet".
	Theme  string // dark|light
	Accent string // default|teal|violet

	// ActiveWindow is how far back in time the "live activity" view reaches
	// (spec §6 [retention] active_window). Clamped to [1h, 168h].
	ActiveWindow time.Duration

	// ContextWindowTokens is the denominator for the Context-Budget tax %. Default
	// 200000 (the current Claude context window); override via [context] window_tokens.
	ContextWindowTokens int

	// Roots is the list of project directories to watch.
	Roots []string

	// AggregateHorizon is how far back aggregate stats (usage/cost) are kept
	// (spec §6 [retention] aggregate_horizon). Zero means keep default.
	AggregateHorizon time.Duration

	// BackupsEnabled controls whether the SQLite WAL is periodically backed up.
	BackupsEnabled bool

	// BackupDir is the directory where backups are written (default: DataDir/backups).
	BackupDir string

	// BackupHorizon is how long backup files are retained before deletion.
	BackupHorizon time.Duration

	// Live source toggles (spec §6 [live]).
	// WatcherEnabled enables the filesystem watcher source.
	WatcherEnabled bool
	// HookPushEnabled enables the stdin hook-push source.
	HookPushEnabled bool
	// StatuslineQuota shows per-session token quota in the Claude statusline.
	StatuslineQuota bool

	// Update check (spec §6 [update]). AutoCheck is the only opt-in outbound call
	// Drishti ever makes; it hits the GitHub releases API for a version string.
	AutoCheck     bool
	CheckInterval time.Duration

	// ThrottleInterval is the scheduler tick interval ([daemon] throttle).
	// Controls how often the background daemon wakes to process queued work.
	ThrottleInterval time.Duration
}

// Default returns the built-in configuration used when no file is present.
// Every field that has a sensible non-zero value is set here; Load then
// overlays values from the TOML file on top.
func Default() Config {
	return Config{
		Port:                7777,
		BindAddr:            "127.0.0.1",
		DataDir:             "", // resolved by Load to the passed dir
		LogLevel:            "info",
		Theme:               "dark",
		Accent:              "default",
		ActiveWindow:        24 * time.Hour,
		ContextWindowTokens: 200000,
		AggregateHorizon:    90 * 24 * time.Hour,
		BackupsEnabled:      true,
		BackupHorizon:       365 * 24 * time.Hour,
		WatcherEnabled:      true,
		HookPushEnabled:     false,
		StatuslineQuota:     false,
		AutoCheck:           false,
		CheckInterval:       24 * time.Hour,
		ThrottleInterval:    time.Second,
	}
}

// Load reads <dir>/config.toml. dir is the resolved data directory.
// Load never returns a hard error for a missing or invalid config file —
// it falls back to defaults and records a Warning instead.
func Load(dir string) (Config, []Warning, error) {
	cfg := Default()
	cfg.DataDir = dir
	var warns []Warning

	path := filepath.Join(dir, "config.toml")
	b, err := os.ReadFile(path)
	if err != nil {
		// Missing file is the normal case on first run: return defaults + a friendly note.
		warns = append(warns, Warning(fmt.Sprintf("no config at %s; using defaults", path)))
		return cfg, warns, nil
	}

	var f fileShape
	if _, derr := toml.Decode(string(b), &f); derr != nil {
		// A totally unparseable file is still non-fatal: log and continue with defaults.
		warns = append(warns, Warning(fmt.Sprintf("invalid config (%v); using defaults", derr)))
		return cfg, warns, nil
	}

	// --- core network ----------------------------------------------------------
	if f.Port != 0 {
		cfg.Port = f.Port
	}
	if f.BindAddr != "" {
		cfg.BindAddr = f.BindAddr
	}
	if f.LogLevel != "" {
		cfg.LogLevel = f.LogLevel
	}

	// --- [ui] ------------------------------------------------------------------
	// Validate Theme against the allowed set {dark, light}.
	// An unrecognised value keeps the default ("dark") and records a Warning.
	if f.UI.Theme != "" {
		switch f.UI.Theme {
		case "dark", "light":
			cfg.Theme = f.UI.Theme
		default:
			warns = append(warns, Warning(fmt.Sprintf("invalid theme %q; using %q", f.UI.Theme, cfg.Theme)))
		}
	}
	// Validate Accent against the allowed set {default, teal, violet}.
	if f.UI.Accent != "" {
		switch f.UI.Accent {
		case "default", "teal", "violet":
			cfg.Accent = f.UI.Accent
		default:
			warns = append(warns, Warning(fmt.Sprintf("invalid accent %q; using %q", f.UI.Accent, cfg.Accent)))
		}
	}

	// --- [retention] -----------------------------------------------------------
	if f.Retention.ActiveWindow != "" {
		if d, perr := time.ParseDuration(f.Retention.ActiveWindow); perr == nil {
			cfg.ActiveWindow = clampWindow(d, &warns)
		} else {
			warns = append(warns, Warning("invalid active_window; using 24h"))
		}
	}
	if f.Retention.AggregateHorizon != "" {
		if d, perr := time.ParseDuration(f.Retention.AggregateHorizon); perr == nil {
			cfg.AggregateHorizon = d
		} else {
			warns = append(warns, Warning("invalid aggregate_horizon; using default"))
		}
	}
	// BackupsEnabled is a bool — a false in the file is intentional, but we only
	// set it if the TOML key was explicitly present (non-zero = true or "was set").
	// Because Go's zero-value for bool is false, we use a pointer in fileShape to
	// distinguish "absent" from "explicitly false".
	if f.Retention.BackupsEnabled != nil {
		cfg.BackupsEnabled = *f.Retention.BackupsEnabled
	}
	if f.Retention.BackupDir != "" {
		cfg.BackupDir = f.Retention.BackupDir
	}
	if f.Retention.BackupHorizon != "" {
		if d, perr := time.ParseDuration(f.Retention.BackupHorizon); perr == nil {
			cfg.BackupHorizon = d
		} else {
			warns = append(warns, Warning("invalid backup_horizon; using default"))
		}
	}

	// --- [context] -------------------------------------------------------------
	if f.Context.WindowTokens > 0 {
		cfg.ContextWindowTokens = f.Context.WindowTokens
	} else if f.Context.WindowTokens < 0 {
		warns = append(warns, Warning("invalid context window_tokens; using 200000"))
	}

	// --- [roots] ---------------------------------------------------------------
	for _, p := range f.Roots.Paths {
		cfg.Roots = append(cfg.Roots, expandHome(p))
	}

	// --- [live] ----------------------------------------------------------------
	if f.Live.WatcherEnabled != nil {
		cfg.WatcherEnabled = *f.Live.WatcherEnabled
	}
	if f.Live.HookPushEnabled != nil {
		cfg.HookPushEnabled = *f.Live.HookPushEnabled
	}
	if f.Live.StatuslineQuota != nil {
		cfg.StatuslineQuota = *f.Live.StatuslineQuota
	}

	// --- [update] --------------------------------------------------------------
	if f.Update.AutoCheck != nil {
		cfg.AutoCheck = *f.Update.AutoCheck
	}
	if f.Update.CheckInterval != "" {
		if d, perr := time.ParseDuration(f.Update.CheckInterval); perr == nil {
			cfg.CheckInterval = d
		} else {
			warns = append(warns, Warning("invalid check_interval; using 24h"))
		}
	}

	// --- [daemon] --------------------------------------------------------------
	if f.Daemon.Throttle != "" {
		if d, perr := time.ParseDuration(f.Daemon.Throttle); perr == nil {
			cfg.ThrottleInterval = d
		} else {
			warns = append(warns, Warning("invalid daemon throttle; using 1s"))
		}
	}

	return cfg, warns, nil
}

// Save writes cfg to <cfg.DataDir>/config.toml. It marshals the full fileShape
// (the typed mirror of the file) and writes atomically: a temp file in the same
// directory then an os.Rename, so a crash mid-write never leaves a half-file.
//
// config.toml ships no comments and is not created on first run, so a full
// rewrite from the typed model loses nothing (see the design spec §7.7).
func Save(cfg Config) error {
	// Project the in-memory Config to the on-disk shape.
	f := toFileShape(cfg)

	// Encode to a buffer in memory first, so we don't create the temp file until
	// we know encoding succeeded.
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(f); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	dir := cfg.DataDir

	// Create a temp file in the same directory as the target. Same-directory is
	// important: os.Rename is atomic only when src and dst are on the same
	// filesystem (and hence the same mount point / directory satisfies that).
	tmp, err := os.CreateTemp(dir, "config-*.toml.tmp")
	if err != nil {
		return fmt.Errorf("temp config: %w", err)
	}
	tmpName := tmp.Name()

	// If anything goes wrong after this point, we must clean up the temp file.
	if _, err := tmp.Write(buf.Bytes()); err != nil {
		tmp.Close()        //nolint:errcheck // already in error path
		os.Remove(tmpName) //nolint:errcheck // best-effort cleanup
		return fmt.Errorf("write config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName) //nolint:errcheck // best-effort cleanup
		return fmt.Errorf("close config: %w", err)
	}

	// Atomic rename: the kernel makes this appear instantaneous to readers.
	// Any process reading config.toml will see either the old or the new file,
	// never a half-written one.
	return os.Rename(tmpName, filepath.Join(dir, "config.toml"))
}

// toFileShape projects the in-memory Config back onto the on-disk TOML shape.
// This is the inverse of the Load decode block above. Every field that Load
// reads must be serialised here, otherwise Save→Load would be lossy.
func toFileShape(cfg Config) fileShape {
	var f fileShape

	// Core network.
	f.Port = cfg.Port
	f.BindAddr = cfg.BindAddr
	f.LogLevel = cfg.LogLevel

	// [ui]
	f.UI.Theme = cfg.Theme
	f.UI.Accent = cfg.Accent

	// [retention] — durations are stored as human-readable strings ("48h0m0s").
	// time.Duration.String() always produces a value that time.ParseDuration
	// accepts back, so the round-trip is lossless.
	f.Retention.ActiveWindow = cfg.ActiveWindow.String()
	f.Retention.AggregateHorizon = cfg.AggregateHorizon.String()
	// Booleans need pointer values so Load can distinguish "absent" from "false".
	backupsEnabled := cfg.BackupsEnabled
	f.Retention.BackupsEnabled = &backupsEnabled
	f.Retention.BackupDir = cfg.BackupDir
	f.Retention.BackupHorizon = cfg.BackupHorizon.String()

	// [context]
	f.Context.WindowTokens = cfg.ContextWindowTokens

	// [roots]
	f.Roots.Paths = cfg.Roots

	// [live] — booleans as pointers (same reason as retention above).
	watcherEnabled := cfg.WatcherEnabled
	f.Live.WatcherEnabled = &watcherEnabled
	hookPushEnabled := cfg.HookPushEnabled
	f.Live.HookPushEnabled = &hookPushEnabled
	statuslineQuota := cfg.StatuslineQuota
	f.Live.StatuslineQuota = &statuslineQuota

	// [update]
	autoCheck := cfg.AutoCheck
	f.Update.AutoCheck = &autoCheck
	f.Update.CheckInterval = cfg.CheckInterval.String()

	// [daemon]
	f.Daemon.Throttle = cfg.ThrottleInterval.String()

	return f
}

// fileShape mirrors the TOML on disk. Zero values mean "unset; keep default".
// Bool fields that need to distinguish "unset" from "explicitly false" use *bool
// (a nil pointer means absent; a non-nil pointer means "user set this").
type fileShape struct {
	Port     int    `toml:"port"`
	BindAddr string `toml:"bind_addr"`
	LogLevel string `toml:"log_level"`

	// [ui]
	UI struct {
		Theme  string `toml:"theme"`
		Accent string `toml:"accent"`
	} `toml:"ui"`

	// [retention]
	Retention struct {
		// ActiveWindow is a duration string, e.g. "48h". Clamped to [1h, 168h].
		ActiveWindow string `toml:"active_window"`
		// AggregateHorizon is how long usage/cost aggregates are kept, e.g. "2160h" (90 days).
		AggregateHorizon string `toml:"aggregate_horizon"`
		// BackupsEnabled controls periodic WAL backups. *bool lets us detect "not set".
		BackupsEnabled *bool `toml:"backups_enabled"`
		// BackupDir overrides the directory where backups are written.
		BackupDir string `toml:"backup_dir"`
		// BackupHorizon is the retention window for backup files, e.g. "8760h" (365 days).
		BackupHorizon string `toml:"backup_horizon"`
	} `toml:"retention"`

	// [context]
	Context struct {
		WindowTokens int `toml:"window_tokens"`
	} `toml:"context"`

	// [roots]
	Roots struct {
		Paths []string `toml:"paths"`
	} `toml:"roots"`

	// [live]
	Live struct {
		// WatcherEnabled enables the filesystem watcher source.
		WatcherEnabled *bool `toml:"watcher_enabled"`
		// HookPushEnabled enables stdin hook-push ingestion.
		HookPushEnabled *bool `toml:"hook_push_enabled"`
		// StatuslineQuota shows per-session token quota in the Claude statusline.
		StatuslineQuota *bool `toml:"statusline_quota"`
	} `toml:"live"`

	// [update]
	Update struct {
		// AutoCheck triggers a version check against the GitHub releases API.
		AutoCheck *bool `toml:"auto_check"`
		// CheckInterval is how often to check, e.g. "24h".
		CheckInterval string `toml:"check_interval"`
	} `toml:"update"`

	// [daemon]
	Daemon struct {
		// Throttle is the scheduler tick interval, e.g. "1s".
		Throttle string `toml:"throttle"`
	} `toml:"daemon"`
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(p string) string {
	if strings.HasPrefix(p, "~") {
		if h, err := os.UserHomeDir(); err == nil {
			return filepath.Join(h, p[1:])
		}
	}
	return p
}

// ClampWindow enforces the [1h, 168h] active-window range from spec §6/D6.
// It returns the (possibly adjusted) duration and any non-fatal Warnings that
// describe what was clamped. It is exported so that the settings package can
// reuse this logic without duplication when validating API input.
func ClampWindow(d time.Duration) (time.Duration, []Warning) {
	var warns []Warning
	result := clampWindow(d, &warns)
	return result, warns
}

// clampWindow is the unexported implementation shared by Load and ClampWindow.
func clampWindow(d time.Duration, warns *[]Warning) time.Duration {
	const minWin, maxWin = time.Hour, 168 * time.Hour
	switch {
	case d < minWin:
		*warns = append(*warns, Warning("active_window below 1h; clamped to 1h"))
		return minWin
	case d > maxWin:
		*warns = append(*warns, Warning("active_window above 168h; clamped to 168h"))
		return maxWin
	default:
		return d
	}
}
