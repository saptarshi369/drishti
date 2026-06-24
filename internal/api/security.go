// Package api — Module 5 Security & Audit handler.
package api

import (
	"net/http"

	"github.com/saptarshi369/drishti/internal/services"
	"github.com/saptarshi369/drishti/internal/store"
)

// handleSecurity serves the persisted security findings as a grouped snapshot.
// Query params (optional): agent (defaulted by agentParam; unknown → typed 400),
// root (default the server's primary root). Empty findings is a normal 200
// "all-clear" snapshot.
func (s *Server) handleSecurity(w http.ResponseWriter, r *http.Request) {
	// v1 is Claude-only; reject unknown agents up front with a typed 400.
	if _, ok := store.AgentID(agentParam(r)); !ok {
		apiError(w, http.StatusBadRequest, "security_invalid", "unknown agent", false)
		return
	}

	// Use the server's configured default root unless the caller overrides with
	// ?root=. Empty string means user-global (no project root).
	root := s.currentDefaultRoot()
	if r.URL.Query().Has("root") {
		root = r.URL.Query().Get("root")
	}

	// Read findings from the store. A DB error here is a 500; an empty result
	// is a normal all-clear 200 (not an error).
	findings, err := s.st.ListFindings("claude", root)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "security_failed", "could not read findings", true)
		return
	}

	// BuildSecuritySnapshot guarantees non-nil Findings and Counts, so the
	// JSON shape is always stable ([] not null, {} not null).
	writeJSON(w, http.StatusOK, services.BuildSecuritySnapshot(findings))
}
