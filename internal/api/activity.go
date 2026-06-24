package api

import (
	"net/http"
	"strconv"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/services"
)

// handleActivity serves the Live Activity snapshot (counters + stream + skills +
// sparklines) for the server's default project root.
//
// # How it works (learner note)
//
// It delegates to services.ActivitySnapshot, which reads several store tables
// and assembles a model.ActivitySnapshot. The handler's only job is to:
//  1. Call the service — if it errors, return a 500 error envelope (spec §11).
//  2. Convert nil slices to empty slices so JSON serialises as [] not null.
//     (The UI iterates these without nil checks; null would cause a runtime error.)
//  3. Write the snapshot as JSON with a 200 status.
//
// Empty slices: in Go, a nil slice and an empty slice behave the same in most
// code, but encoding/json marshals a nil slice as null and an empty slice as
// []. We always want [], so we replace nil with an initialised empty slice.
func (s *Server) handleActivity(w http.ResponseWriter, _ *http.Request) {
	snap, err := services.ActivitySnapshot(s.st, s.currentDefaultRoot())
	if err != nil {
		apiError(w, http.StatusInternalServerError, "activity_failed", "could not assemble activity", true)
		return
	}

	// Guarantee [] not null in JSON output. nil slices are valid Go but they
	// marshal as JSON null, which breaks UI iteration without nil guards.
	if snap.Recent == nil {
		snap.Recent = []model.RecentEvent{}
	}
	if snap.Skills == nil {
		snap.Skills = []model.SkillTrigger{}
	}

	writeJSON(w, http.StatusOK, snap)
}

// handleActivityEvents serves a newest-first page of events for stream scroll-back.
// Params: type (optional code filter), limit (default 50, max 200), before (id cursor).
//
// Keyset pagination (Go learner note):
//
//	The UI passes the database id of the last row it already rendered as the
//	`before` parameter. The store returns only rows with id < before, so the
//	next page never re-fetches rows the UI already has. before=0 (or absent)
//	means "start from the newest row".
//
// Query-param parsing: strconv.Atoi/ParseInt return an error on bad input; we
// ignore those errors and treat the result as 0, which the store normalises:
// limit=0 → 50 (default), before=0 → no cursor (start from newest).
func (s *Server) handleActivityEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	// Parse limit; ignore parse errors — 0 triggers the store's default clamp.
	limit, _ := strconv.Atoi(q.Get("limit"))
	// ParseInt with bitSize=64 to safely hold a SQLite int64 row id.
	before, _ := strconv.ParseInt(q.Get("before"), 10, 64)
	evs, err := s.st.EventsPage(q.Get("type"), limit, before)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "events_failed", "could not read events", true)
		return
	}
	// Convert nil to [] so JSON serialises as [] not null. The UI iterates this
	// slice without nil checks, so null would cause a runtime error in the frontend.
	if evs == nil {
		evs = []model.RecentEvent{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": evs})
}
