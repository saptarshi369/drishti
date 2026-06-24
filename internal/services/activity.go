// Package services assembles higher-level payloads from store reads and any
// derived computations (e.g. cost back-fill). It is the layer between the raw
// SQL results in store and the JSON shapes the HTTP/SSE layer serves. Services
// keep HTTP concerns out of the store and keep SQL out of the API handlers.
package services

import (
	"time"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/store"
)

// sparkBuckets is the number of per-minute samples each sparkline carries.
// 30 gives a half-hour rolling window at one-minute granularity — enough to
// show rhythm without overwhelming the header. Downstream callers (Task 10/12)
// get a fixed-length slice; callers should not hard-code 30 themselves.
const sparkBuckets = 30

// ActivitySnapshot assembles the complete Live Activity payload:
//
//   - Counters.Last24h — tallies for all events in the last 24 hours (any session).
//   - Counters.Session — tallies for the most-recently-started session only
//     (skipped and left as zero-value if no sessions exist yet).
//   - Recent         — the 40 most-recent events (newest first) for the live ticker.
//   - Skills         — active inventory skills with cumulative trigger counts +
//     Dead flag (never-fired skills flagged as archive candidates).
//   - Sparklines     — 30-minute prompts/min and skills/min series for the header graph.
//
// It is the single source both the /api/activity REST handler (Task 10) and the
// SSE broadcaster (Task 12) use, so the page and stream never disagree.
//
// Every store error short-circuits and returns a zero ActivitySnapshot. The
// caller decides how to handle the error; we never panic (spec §14 / failsafe).
func ActivitySnapshot(st *store.Store, projectRoot string) (model.ActivitySnapshot, error) {
	// Anchor all time-relative queries to a single "now" so every field in the
	// snapshot is coherent (no sub-millisecond drift between calls).
	now := time.Now().UnixMilli()

	// dayAgo is the Unix-millisecond lower bound for the 24h rolling window.
	// time.Hour is a time.Duration (nanoseconds); dividing by time.Millisecond
	// converts it to milliseconds. Multiplying by 24 gives 86_400_000 ms.
	dayAgo := now - 24*int64(time.Hour/time.Millisecond)

	// Tally all events in the last 24 hours, across all sessions.
	// Passing "" as sessionID means "no session filter" (see ActivityCounters docs).
	last24, err := st.ActivityCounters(dayAgo, "")
	if err != nil {
		return model.ActivitySnapshot{}, err
	}

	// Find the most-recently-started session to anchor the "current session" pane.
	// LatestSession returns ("", 0, nil) when no sessions exist — that is fine;
	// we guard below and leave session as a zero CounterSet in that case.
	sessID, _, err := st.LatestSession()
	if err != nil {
		return model.ActivitySnapshot{}, err
	}

	// Only query session counters when there is an actual session to look up.
	// Querying with sessID=="" would return all-sessions data (same as Last24h with
	// sinceMs=0), which would be misleading — so we skip the call and keep the
	// zero CounterSet the UI renders as "—" when no session has started yet.
	var session model.CounterSet
	if sessID != "" {
		// sinceMs=0 means "all time", restricted by session id only.
		if session, err = st.ActivityCounters(0, sessID); err != nil {
			return model.ActivitySnapshot{}, err
		}
	}

	// Fetch the 40 most-recent events for the live ticker (newest first from the store).
	recent, err := st.RecentEvents(40)
	if err != nil {
		return model.ActivitySnapshot{}, err
	}

	// Active inventory skills with cumulative trigger counts. The store LEFT-JOINs
	// skill_stats so skills that have never fired appear with Count=0, Dead=true.
	skills, err := st.SkillTriggers(projectRoot)
	if err != nil {
		return model.ActivitySnapshot{}, err
	}

	// Per-minute prompt and skill rates for the last sparkBuckets minutes.
	// EventRatePerMinute always returns a slice of exactly sparkBuckets elements
	// (zero-filled for quiet minutes), so PromptsPerMin and SkillsPerMin are
	// always length 30 — the UI can render them without length checks.
	prompts, err := st.EventRatePerMinute("prompt", now, sparkBuckets)
	if err != nil {
		return model.ActivitySnapshot{}, err
	}
	skillSpark, err := st.EventRatePerMinute("skill", now, sparkBuckets)
	if err != nil {
		return model.ActivitySnapshot{}, err
	}

	return model.ActivitySnapshot{
		Counters: model.ActivityCounters{
			Session: session,
			Last24h: last24,
		},
		Recent: recent,
		Skills: skills,
		Sparklines: model.Sparklines{
			PromptsPerMin: prompts,
			SkillsPerMin:  skillSpark,
		},
	}, nil
}
