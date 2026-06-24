// Package api provides the HTTP server, SSE hub, and route handlers for Drishti.
// This file adds the inventory handlers (spec §16).
package api

import (
	"net/http"
	"strconv"

	"github.com/saptarshi369/drishti/internal/model"
)

// handleInventory serves the resolved inventory rows from the store's
// materialized view. Query params (all optional):
//
//   - category — filter by "skill", "agent", "hook", or "mcp"; empty = all
//   - root     — filter by project root path; empty = all roots
//   - show_disabled — "1" to include disabled/shadowed rows (default: active only)
//
// Response: {"items": [ResolvedRow...]} — items is always a JSON array,
// never null, so the UI can iterate without a nil check.
func (s *Server) handleInventory(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	// When the request omits ?root=, use the server's default root (the daemon's
	// primary/~ project root), whose resolved set merges user + project scope.
	// An explicit ?root= (including empty) is honoured as-is.
	root := s.currentDefaultRoot()
	if q.Has("root") {
		root = q.Get("root")
	}
	rows, err := s.st.ListResolved(q.Get("category"), root, q.Get("show_disabled") == "1")
	if err != nil {
		// Wrap the DB error in the uniform error envelope (spec §11); never
		// expose the raw error to callers.
		apiError(w, http.StatusInternalServerError, "inventory_failed", "could not read inventory", true)
		return
	}
	// Ensure the JSON array is [] rather than null when there are no rows.
	// Go's json.Encode serialises a nil slice as "null"; an empty slice as "[]".
	if rows == nil {
		rows = []model.ResolvedRow{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": rows})
}

// handleInventoryWhy serves the precedence trail for one resolved row (the
// "why?" drawer in the UI). The {id} path value is the integer primary key
// from the resolved_inventory table.
//
// Returns {"trail": [PrecedenceStep...]} on success, or a 400 if the id is
// not a valid integer, or a 500 if the DB read fails.
func (s *Server) handleInventoryWhy(w http.ResponseWriter, r *http.Request) {
	// r.PathValue("id") is available in Go 1.22+ ServeMux with pattern {id}.
	raw := r.PathValue("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		// A non-numeric path segment is a client mistake, not a server error.
		apiError(w, http.StatusBadRequest, "bad_id", "invalid inventory id", false)
		return
	}
	trail, err := s.st.ResolvedTrail(id)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "trail_failed", "could not read precedence trail", true)
		return
	}
	// trail may be nil (unknown id) — return an empty array for consistency.
	if trail == nil {
		trail = []model.PrecedenceStep{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"trail": trail})
}
