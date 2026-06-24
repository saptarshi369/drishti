package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/saptarshi369/drishti/internal/services"
	"github.com/saptarshi369/drishti/internal/updater"
)

// startedAt anchors uptime reporting.
var startedAt = time.Now()

// writeJSON encodes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// apiError is the uniform error envelope (spec §11): never a bare 500.
func apiError(w http.ResponseWriter, code int, errCode, msg string, retryable bool) {
	writeJSON(w, code, map[string]any{
		"error": map[string]any{"code": errCode, "message": msg, "retryable": retryable},
	})
}

// handleHealth reports liveness.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"uptime_ms": time.Since(startedAt).Milliseconds(),
		"version":   s.version,
	})
}

// handleOverview returns the full Overview snapshot (KPIs + recent + M8 aggregates).
func (s *Server) handleOverview(w http.ResponseWriter, _ *http.Request) {
	ov, err := services.OverviewSnapshot(s.st, s.overviewParams())
	if err != nil {
		apiError(w, http.StatusInternalServerError, "overview_failed", "could not assemble overview", true)
		return
	}
	writeJSON(w, http.StatusOK, ov)
}

// overviewParams snapshots the mutex-guarded server settings the Overview
// assembler needs, via the M7 accessors (never the raw fields — data-race guard).
func (s *Server) overviewParams() services.OverviewParams {
	root := s.currentDefaultRoot()
	return services.OverviewParams{
		Root:            root,
		ProjectKey:      services.EncodeProjectKey(root),
		WindowTokens:    s.currentContextWindowTokens(),
		SkillThresholds: s.currentSkillThresholds(),
	}
}

// handleUpdateStatus reports the updater status. When the user has opted in
// (auto_check=true in config) or sends ?check=1, it performs a live GitHub
// Releases query via updater.Check. Otherwise it returns CurrentStatus with
// no network call. Any failure degrades silently to CurrentStatus (spec §14).
func (s *Server) handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	// Snapshot cfg once to avoid a TOCTOU race: the scheduler may call
	// SetConfig concurrently, and cfg.AutoCheck is a bool that could be
	// partially updated on some architectures without a lock.
	cfg := s.snapshotConfig()
	// Opt-in gate: either the persistent config flag or a one-shot query param.
	// This is the ONLY place in Drishti that initiates outbound network I/O.
	if cfg.AutoCheck || r.URL.Query().Get("check") == "1" {
		// Bound the outbound call to 5 seconds so a slow GitHub API never
		// hangs the UI. WithTimeout returns a cancel func; defer ensures it
		// always runs so the timer goroutine is released (Go best practice).
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		writeJSON(w, http.StatusOK, updater.Check(ctx, s.version, http.DefaultClient))
		return
	}
	// No opt-in: return the stub status; no network involved.
	writeJSON(w, http.StatusOK, updater.CurrentStatus(s.version))
}
