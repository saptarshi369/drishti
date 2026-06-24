// Package api — top-bar active-root selector. The single global "view root"
// (currentDefaultRoot) scopes every screen: the SSE-driven Overview/Activity and
// the inventory/context/skills/security pages. GET lists the selectable roots and
// the current selection; PUT switches it. Selection is in-memory (a daemon restart
// resets to the configured primary root); it is NOT a config change, so it never
// touches the watched-roots list in config.toml.
package api

import (
	"encoding/json"
	"net/http"
	"os"
)

// rootOptions returns the home directory and the de-duplicated list of roots the
// user may scope to: home (the user-global ~ view) followed by every configured
// watched root. Home is included even when roots are configured so the user can
// always return to the global view.
func (s *Server) rootOptions() (home string, opts []string) {
	home, _ = os.UserHomeDir()
	cfg := s.snapshotConfig()
	seen := map[string]bool{}
	add := func(p string) {
		if p != "" && !seen[p] {
			seen[p] = true
			opts = append(opts, p)
		}
	}
	add(home)
	for _, r := range cfg.Roots {
		add(r)
	}
	return home, opts
}

// handleGetActiveRoot serves GET /api/active-root → the selectable roots and the
// current selection, so the top-bar dropdown can render its options and highlight
// the active one.
func (s *Server) handleGetActiveRoot(w http.ResponseWriter, _ *http.Request) {
	home, opts := s.rootOptions()
	writeJSON(w, http.StatusOK, map[string]any{
		"current": s.currentDefaultRoot(),
		"default": s.currentPrimaryRoot(),
		"home":    home,
		"roots":   opts,
	})
}

// handleSetActiveRoot serves PUT /api/active-root {"root": "..."}. It validates the
// requested root is one of the current options (home or a configured root), sets it
// as the global view root, and pushes a fresh snapshot so the SSE-driven Overview
// re-scopes immediately instead of waiting for the next heartbeat. An unknown root
// is a typed 400 and leaves the current scope unchanged.
func (s *Server) handleSetActiveRoot(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Root string `json:"root"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiError(w, http.StatusBadRequest, "active_root_invalid", "malformed body", false)
		return
	}
	_, opts := s.rootOptions()
	valid := false
	for _, o := range opts {
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
