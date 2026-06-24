package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/saptarshi369/drishti/internal/config"
)

// newSettingsTestServer builds a nil-store Server with a temp DataDir for the
// settings handler tests. Settings handlers don't touch the store, so nil is fine.
func newSettingsTestServer(t *testing.T) *Server {
	t.Helper()
	s := NewServer("test", nil)
	cfg := config.Default()
	cfg.DataDir = t.TempDir()
	s.SetConfig(cfg)
	return s
}

func TestPutSettingsValidatesPort(t *testing.T) {
	s := newSettingsTestServer(t)
	body, _ := json.Marshal(map[string]any{"port": 70000})
	r := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.handlePutSettings(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestPutSettingsRestartRequiredFlag(t *testing.T) {
	s := newSettingsTestServer(t)
	body, _ := json.Marshal(map[string]any{"port": 9001})
	r := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.handlePutSettings(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp struct {
		Saved           bool `json:"saved"`
		RestartRequired bool `json:"restart_required"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Saved || !resp.RestartRequired {
		t.Errorf("resp = %+v, want saved+restart_required", resp)
	}
}

// TestSettingsRoutesRegistered verifies that /api/settings is wired into the
// router — i.e., Handler() returns 200 not 404 for GET and 400 not 404 for PUT.
func TestSettingsRoutesRegistered(t *testing.T) {
	s := newSettingsTestServer(t)
	h := s.Handler()

	// GET /api/settings must return 200 (not 404).
	r := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("GET /api/settings: status = %d, want 200", w.Code)
	}
}

// TestGetSettingsIncludesDiskBytes verifies that GET /api/settings returns
// db_bytes and backup_bytes fields (disk estimate) in the JSON payload.
// With an empty temp DataDir the values are 0; we assert the keys are present
// in the raw JSON body.
func TestGetSettingsIncludesDiskBytes(t *testing.T) {
	s := newSettingsTestServer(t)
	r := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	w := httptest.NewRecorder()
	s.handleGetSettings(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"db_bytes"`) {
		t.Errorf("response missing db_bytes field; body: %s", body)
	}
	if !strings.Contains(body, `"backup_bytes"`) {
		t.Errorf("response missing backup_bytes field; body: %s", body)
	}
}

// TestGetSettingsMCPServersNilStore verifies that GET /api/settings on a
// nil-store server still returns a valid mcp_servers field as a non-nil empty
// array (JSON []), not null. This guards the best-effort nil-safe path added
// for the Agents card (live `claude mcp list` probe is deferred).
func TestGetSettingsMCPServersNilStore(t *testing.T) {
	// newSettingsTestServer uses nil store — exercises the nil-safe guard.
	s := newSettingsTestServer(t)
	r := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	w := httptest.NewRecorder()
	s.handleGetSettings(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp struct {
		MCPServers []string `json:"mcp_servers"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	// MCPServers must be non-nil (JSON [] not null) and empty for a nil-store server.
	if resp.MCPServers == nil {
		t.Error("mcp_servers is null in JSON; want non-nil empty array []")
	}
	if len(resp.MCPServers) != 0 {
		t.Errorf("mcp_servers len = %d, want 0 for nil-store server", len(resp.MCPServers))
	}
}
