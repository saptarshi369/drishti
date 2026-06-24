// Package api — top-bar active-root selector. The single global "view scope"
// (currentDefaultRoot) filters every screen: the SSE-driven Overview/Activity and
// the inventory/context/skills/security pages. The empty string is the "All" scope
// (user-global inventory + no usage/event filter) and is the default; a non-empty
// value is one configured folder. GET lists the folders + current selection; PUT
// switches it. Selection is in-memory (a daemon restart resets to All); it is NOT a
// config change, so it never touches the watched-roots list in config.toml.
package api

import (
	"encoding/json"
	"net/http"
)

// rootOptions returns the de-duplicated list of configured watched folders the
// user may scope to. The "All" scope (empty string, the default) is implicit — the
// client renders it as the first option; it is not a configured root.
func (s *Server) rootOptions() []string {
	cfg := s.snapshotConfig()
	seen := map[string]bool{}
	opts := []string{}
	for _, r := range cfg.Roots {
		if r != "" && !seen[r] {
			seen[r] = true
			opts = append(opts, r)
		}
	}
	return opts
}

// handleGetActiveRoot serves GET /api/active-root → the selectable folders and the
// current selection ("" = All), so the top-bar dropdown can render "All" + each
// folder and highlight the active one.
func (s *Server) handleGetActiveRoot(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"current": s.currentDefaultRoot(), // "" = All (the default scope)
		"default": s.currentPrimaryRoot(), // daemon's primary configured folder (info)
		"roots":   s.rootOptions(),
	})
}

// handleSetActiveRoot serves PUT /api/active-root {"root": "..."}. The empty string
// selects "All" (no filter); otherwise the root must be one of the configured
// folders. It sets the global view scope and pushes a fresh snapshot so the
// SSE-driven Overview re-scopes immediately instead of waiting for the next
// heartbeat. An unknown root is a typed 400 and leaves the current scope unchanged.
func (s *Server) handleSetActiveRoot(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Root string `json:"root"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiError(w, http.StatusBadRequest, "active_root_invalid", "malformed body", false)
		return
	}
	valid := body.Root == "" // "" = All is always selectable
	for _, o := range s.rootOptions() {
		if o == body.Root {
			valid = true
			break
		}
	}
	if !valid {
		apiError(w, http.StatusBadRequest, "active_root_unknown", "root is not a configured root", false)
		return
	}
	s.SetSelectedRoot(body.Root)
	// Re-scope the live Overview/Activity right away (best-effort; nil store in
	// tests simply broadcasts a status frame).
	s.BroadcastSnapshot()
	writeJSON(w, http.StatusOK, map[string]any{"current": body.Root})
}
