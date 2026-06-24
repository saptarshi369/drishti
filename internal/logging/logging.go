// Package logging configures Drishti's structured logger.
//
// Logs are LOCAL ONLY (never network). We write JSON records to a file under the
// data dir. We never log secrets or raw prompt text — scrub before logging, the
// same rule as before storing.
package logging

import (
	"log/slog"
	"os"
	"path/filepath"
)

// New builds a *slog.Logger that writes JSON records to <dir>/logs/drishti.log.
// level is one of debug|info|warn|error (anything else defaults to info).
func New(dir, level string) (*slog.Logger, error) {
	logDir := filepath.Join(dir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(logDir, "drishti.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	h := slog.NewJSONHandler(f, &slog.HandlerOptions{Level: parseLevel(level)})
	return slog.New(h), nil
}

// parseLevel maps a config string to an slog.Level, defaulting to info.
func parseLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
