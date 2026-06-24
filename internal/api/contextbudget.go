// Package api — Module 4 Context-Budget handler.
package api

import (
	"net/http"

	"github.com/saptarshi369/drishti/internal/services"
	"github.com/saptarshi369/drishti/internal/store"
)

// handleContextBudget serves the always-on context-tax snapshot. Query params
// (optional): agent (defaulted by agentParam; unknown → typed 400), root (default the
// server's primary root). Reads the materialized resolved rows and folds them in
// BuildContextBudget. Empty inventory is a normal 200 with a zero snapshot.
func (s *Server) handleContextBudget(w http.ResponseWriter, r *http.Request) {
	// v1 is Claude-only; reject unknown agents up front (mirrors the quota path).
	if _, ok := store.AgentID(agentParam(r)); !ok {
		apiError(w, http.StatusBadRequest, "context_budget_invalid", "unknown agent", false)
		return
	}
	root := s.currentDefaultRoot()
	if r.URL.Query().Has("root") {
		root = r.URL.Query().Get("root")
	}
	// showDisabled=true returns all statuses; BuildContextBudget keeps only active.
	rows, err := s.st.ListResolved("", root, true)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "context_budget_failed", "could not read inventory", true)
		return
	}
	writeJSON(w, http.StatusOK, services.BuildContextBudget(rows, s.currentContextWindowTokens()))
}
