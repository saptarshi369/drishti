package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/services"
	"github.com/saptarshi369/drishti/internal/store"
)

// agentParam returns the ?agent= value or the "claude" default (v1 is Claude-only).
func agentParam(r *http.Request) string {
	if a := r.URL.Query().Get("agent"); a != "" {
		return a
	}
	return "claude"
}

// handleUsage serves the Usage & Cost snapshot (7-day trend, breakdowns, heatmap).
// nil slices are converted to empty slices so JSON serialises as [] not null.
func (s *Server) handleUsage(w http.ResponseWriter, r *http.Request) {
	snap, err := services.UsageSnapshot(s.st, agentParam(r))
	if err != nil {
		apiError(w, http.StatusInternalServerError, "usage_failed", "could not assemble usage", true)
		return
	}
	if snap.Days == nil {
		snap.Days = []model.DailyUsage{}
	}
	if snap.ByProject == nil {
		snap.ByProject = []model.ProjectUsage{}
	}
	if snap.ByModel == nil {
		snap.ByModel = []model.UsageShare{}
	}
	if snap.Heatmap == nil {
		snap.Heatmap = []model.HeatDay{}
	}
	writeJSON(w, http.StatusOK, snap)
}

// handleQuota serves the live plan-quota snapshot. "No samples yet" is a normal
// gated state (Available=false), not an error — so it returns 200, letting the UI
// render the install-helper call to action.
func (s *Server) handleQuota(w http.ResponseWriter, r *http.Request) {
	snap, err := services.QuotaSnapshot(s.st, agentParam(r))
	if err != nil {
		apiError(w, http.StatusInternalServerError, "quota_failed", "could not read quota", true)
		return
	}
	writeJSON(w, http.StatusOK, snap)
}

// quotaSampleBody is the JSON the statusline helper forwards. Windows are
// pointers so we can tell "absent" from "zero".
type quotaSampleBody struct {
	Agent  string `json:"agent"`
	Plan   string `json:"plan"`
	Source string `json:"source"`
	// FiveHour is the 5-hour rolling window quota reading; nil means absent.
	FiveHour *struct {
		UsedPercentage float64 `json:"used_percentage"`
		ResetsAtMs     int64   `json:"resets_at_ms"`
	} `json:"five_hour"`
	// SevenDay is the 7-day rolling window quota reading; nil means absent.
	SevenDay *struct {
		UsedPercentage float64 `json:"used_percentage"`
		ResetsAtMs     int64   `json:"resets_at_ms"`
	} `json:"seven_day"`
}

// handleQuotaSample ingests one forwarded quota reading (one row per present
// window) and broadcasts a fresh "quota" SSE frame. A body with no usable window
// is a typed 400 (degrade, never a 500). Localhost-only by virtue of the daemon's
// bind address; no auth in v1 (zero outbound network, single-user local tool).
func (s *Server) handleQuotaSample(w http.ResponseWriter, r *http.Request) {
	var b quotaSampleBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		apiError(w, http.StatusBadRequest, "quota_sample_invalid", "malformed quota sample", false)
		return
	}
	agent := b.Agent
	if agent == "" {
		agent = "claude"
	}
	// Reject unknown agents with a typed 400 before any DB write. v1 is
	// Claude-only; falling back silently to claude would mis-attribute data
	// from future agents (e.g. codex). store.AgentID delegates to the unexported
	// agentID mapping so the two places never get out of sync.
	if _, ok := store.AgentID(agent); !ok {
		apiError(w, http.StatusBadRequest, "quota_sample_invalid", "unknown agent", false)
		return
	}
	source := b.Source
	if source == "" {
		source = "statusline"
	}
	// Use wall-clock millis for TsMs so the row sorts correctly even if the
	// helper omits an explicit timestamp (the sample is "now" from the daemon's
	// perspective).
	now := time.Now().UnixMilli()

	wrote := 0
	if b.FiveHour != nil {
		if err := s.st.InsertQuotaSample(model.QuotaSampleRow{
			AgentCode:      agent,
			Window:         "five_hour",
			UsedPercentage: b.FiveHour.UsedPercentage,
			ResetsAtMs:     b.FiveHour.ResetsAtMs,
			TsMs:           now,
			Plan:           b.Plan,
			Source:         source,
		}); err != nil {
			apiError(w, http.StatusInternalServerError, "quota_store_failed", "could not store sample", true)
			return
		}
		wrote++
	}
	if b.SevenDay != nil {
		if err := s.st.InsertQuotaSample(model.QuotaSampleRow{
			AgentCode:      agent,
			Window:         "seven_day",
			UsedPercentage: b.SevenDay.UsedPercentage,
			ResetsAtMs:     b.SevenDay.ResetsAtMs,
			TsMs:           now,
			Plan:           b.Plan,
			Source:         source,
		}); err != nil {
			apiError(w, http.StatusInternalServerError, "quota_store_failed", "could not store sample", true)
			return
		}
		wrote++
	}
	if wrote == 0 {
		apiError(w, http.StatusBadRequest, "quota_sample_invalid", "no quota window in body", false)
		return
	}

	// Broadcast the fresh snapshot so connected gauges update live.
	if snap, err := services.QuotaSnapshot(s.st, agent); err == nil {
		s.hub.Broadcast(Message{Type: "quota", TS: now, Payload: snap})
	}
	w.WriteHeader(http.StatusNoContent)
}
