package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/saptarshi369/drishti/internal/security"
	"github.com/saptarshi369/drishti/internal/skills"
)

// newConfigFilesTestServer builds a nil-store Server with temp file paths set
// via SetConfigFilePaths. The config-file handlers do not touch the store or
// config, so nil is safe here.
func newConfigFilesTestServer(t *testing.T) (*Server, string, string) {
	t.Helper()
	dir := t.TempDir()
	thresholdsPath := filepath.Join(dir, "skills-analytics.toml")
	rulesPath := filepath.Join(dir, "security-rules.toml")
	s := NewServer("test", nil)
	s.SetConfigFilePaths(rulesPath, thresholdsPath)
	return s, rulesPath, thresholdsPath
}

// TestPutThresholds_InvalidBody posts thresholds with HighTriggerMin=0 (invalid per
// Thresholds.Valid()) and expects a typed 400 response.
func TestPutThresholds_InvalidBody(t *testing.T) {
	s, _, _ := newConfigFilesTestServer(t)
	// HighTriggerMin = 0 fails Valid(); LowValueRatioMax alone being positive is
	// not enough — both fields must be positive.
	body, _ := json.Marshal(map[string]any{"high_trigger_min": 0, "low_value_ratio_max": 1.0})
	r := httptest.NewRequest(http.MethodPut, "/api/thresholds", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.handlePutThresholds(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

// TestPutThresholds_ValidBody posts valid thresholds and asserts 200 + round-trip.
func TestPutThresholds_ValidBody(t *testing.T) {
	s, _, thresholdsPath := newConfigFilesTestServer(t)
	body, _ := json.Marshal(skills.Thresholds{HighTriggerMin: 30, LowValueRatioMax: 0.7})
	r := httptest.NewRequest(http.MethodPut, "/api/thresholds", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.handlePutThresholds(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Saved bool `json:"saved"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Saved {
		t.Errorf("saved = false, want true")
	}
	// Verify the file was written with the correct values.
	got := skills.LoadThresholdsFromPath(thresholdsPath, nil)
	if got.HighTriggerMin != 30 || got.LowValueRatioMax != 0.7 {
		t.Errorf("file round-trip: got %+v, want {30 0.7}", got)
	}
}

// TestPutRules_ValidBody posts a valid rules JSON array and asserts 200 + file round-trip.
func TestPutRules_ValidBody(t *testing.T) {
	s, rulesPath, _ := newConfigFilesTestServer(t)
	rules := []security.Rule{
		{
			ID:       "r1",
			Type:     "forbid-mode",
			Enabled:  true,
			Severity: "high",
			Title:    "No bypass",
			Modes:    []string{"bypassPermissions"},
		},
	}
	body, _ := json.Marshal(rules)
	r := httptest.NewRequest(http.MethodPut, "/api/rules", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.handlePutRules(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Saved bool `json:"saved"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Saved {
		t.Errorf("saved = false, want true")
	}
	// Verify the file was written with the correct values.
	got := security.LoadRulesFromPath(rulesPath, nil)
	if len(got) != 1 || got[0].ID != "r1" {
		t.Errorf("file round-trip: got %+v, want [{ID:r1}]", got)
	}
}

// TestPutRules_InvalidBody posts a rule with an empty ID and expects a typed 400.
func TestPutRules_InvalidBody(t *testing.T) {
	s, _, _ := newConfigFilesTestServer(t)
	// An empty ID must be rejected — the UI should never silently discard a rule.
	rules := []security.Rule{
		{ID: "", Type: "forbid-mode", Enabled: true, Severity: "high"},
	}
	body, _ := json.Marshal(rules)
	r := httptest.NewRequest(http.MethodPut, "/api/rules", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.handlePutRules(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

// TestGetThresholds returns the server's current skillThresholds as JSON.
// The test constructs a Server, sets thresholds via SetSkillThresholds, then
// GETs /api/thresholds and asserts 200 + correct decoded values.
func TestGetThresholds(t *testing.T) {
	s, _, _ := newConfigFilesTestServer(t)
	// Set non-default thresholds so we can distinguish them from zero values.
	s.SetSkillThresholds(skills.Thresholds{HighTriggerMin: 30, LowValueRatioMax: 0.7})
	r := httptest.NewRequest(http.MethodGet, "/api/thresholds", nil)
	w := httptest.NewRecorder()
	s.handleGetThresholds(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	var got skills.Thresholds
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.HighTriggerMin != 30 {
		t.Errorf("HighTriggerMin = %d, want 30", got.HighTriggerMin)
	}
	if got.LowValueRatioMax != 0.7 {
		t.Errorf("LowValueRatioMax = %f, want 0.7", got.LowValueRatioMax)
	}
}
