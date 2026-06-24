// Package api — Module 6 Skills Analytics handler.
package api

import (
	"net/http"

	"github.com/saptarshi369/drishti/internal/skills"
	"github.com/saptarshi369/drishti/internal/store"
)

// handleSkills serves per-skill analytics: triggers vs. always-on context cost,
// a value ratio, and the dead / over-triggering / disabled flags. It mirrors
// handleSecurity: claude-only (unknown agent → typed 400), optional ?root=
// override of the server's default root, empty result is a normal 200.
//
// The thresholds come from s.skillThresholds (set/refreshed by the daemon), so
// this handler does no file I/O.
func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	// v1 is Claude-only; reject unknown agents up front with a typed 400.
	if _, ok := store.AgentID(agentParam(r)); !ok {
		apiError(w, http.StatusBadRequest, "skills_invalid", "unknown agent", false)
		return
	}

	// Use the server's configured default root unless ?root= overrides it.
	root := s.currentDefaultRoot()
	if r.URL.Query().Has("root") {
		root = r.URL.Query().Get("root")
	}

	rows, err := s.st.SkillAnalytics(root)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "skills_failed", "could not read skill analytics", true)
		return
	}

	// BuildAnalytics guarantees a non-nil Items slice, so the JSON shape is
	// stable ([] not null) even for the all-clear case.
	writeJSON(w, http.StatusOK, skills.BuildAnalytics(rows, s.currentSkillThresholds()))
}
