// Package api — Module 7 config-file editors. PUT the rule files (security +
// skills) from the parsed model. The ~10s scheduler reload applies them live.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/saptarshi369/drishti/internal/security"
	"github.com/saptarshi369/drishti/internal/skills"
)

// handleGetThresholds returns the server's current skill-analytics thresholds as
// JSON (GET /api/thresholds). The returned object has two fields:
//
//	{"high_trigger_min": 25, "low_value_ratio_max": 0.4}
//
// The values reflect what the daemon loaded most recently from skills-analytics.toml
// (updated every ~10 s by the scheduler). The Settings Config-files panel uses
// this to seed its threshold inputs so the user edits the actual saved values
// rather than stale defaults.
func (s *Server) handleGetThresholds(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.currentSkillThresholds())
}

// handlePutThresholds saves skills-analytics.toml from the posted Thresholds
// value. The request body must be a JSON object with the two threshold fields:
//
//	{"high_trigger_min": 25, "low_value_ratio_max": 0.4}
//
// Both fields must be positive (Thresholds.Valid()); any other values return a
// 400. On success the file is rewritten atomically and the daemon's ~10s
// scheduler will pick it up automatically.
func (s *Server) handlePutThresholds(w http.ResponseWriter, r *http.Request) {
	var t skills.Thresholds
	// json.Decoder.Decode maps JSON keys to struct fields by the standard json
	// tag rules; Thresholds uses toml tags not json tags, so we decode by field
	// name (Go default: field name lowercased). The TOML key names differ
	// (high_trigger_min), so we decode from JSON using the exported field names.
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		apiError(w, http.StatusBadRequest, "thresholds_invalid", "malformed body", false)
		return
	}
	// Valid() checks HighTriggerMin > 0 && LowValueRatioMax > 0. Zero or
	// negative values would silently disable the over-triggering flag, so we
	// reject them with a typed 400 rather than writing a broken file.
	if !t.Valid() {
		apiError(w, http.StatusBadRequest, "thresholds_invalid",
			"high_trigger_min and low_value_ratio_max must both be positive", false)
		return
	}
	if err := skills.WriteThresholds(s.thresholdsPath, t); err != nil {
		apiError(w, http.StatusInternalServerError, "thresholds_save_failed",
			"could not write thresholds file", true)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"saved": true})
}

// handlePutRules saves security-rules.toml from the posted rules array. The
// request body must be a JSON array of rule objects:
//
//	[{"id":"...", "type":"...", "enabled":true, ...}, ...]
//
// Each rule must have a non-empty ID and Type; any rule that fails this check
// causes a 400 and no file is written. On success the file is rewritten
// atomically and the daemon's ~10s scheduler will pick it up automatically.
// Note: the body is a flat JSON array of rule objects (not {"rule":[...]}) —
// this is the simplest shape for the Settings UI to POST.
func (s *Server) handlePutRules(w http.ResponseWriter, r *http.Request) {
	var rules security.Rules
	if err := json.NewDecoder(r.Body).Decode(&rules); err != nil {
		apiError(w, http.StatusBadRequest, "rules_invalid", "malformed body", false)
		return
	}
	// Validate that every rule has a non-empty ID and Type. The engine will skip
	// rules with unknown types, but the Settings UI should not silently discard
	// user edits — reject the whole batch if any entry is obviously broken.
	for i, rule := range rules {
		if rule.ID == "" || rule.Type == "" {
			apiError(w, http.StatusBadRequest, "rules_invalid",
				"each rule must have a non-empty id and type", false)
			_ = i // keep linter happy; the index is implicit in the message
			return
		}
	}
	if err := security.WriteRules(s.rulesPath, rules); err != nil {
		apiError(w, http.StatusInternalServerError, "rules_save_failed",
			"could not write rules file", true)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"saved": true})
}
