// Package api — Module 7 install handler tests.
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestHandleProposeStatusline_WithExisting verifies that GET /api/install/statusline
// returns 200 with a valid proposed JSON that contains statusLine AND preserves
// the caller's existing keys. It also proves non-mutation: the on-disk file is
// byte-identical after the call.
func TestHandleProposeStatusline_WithExisting(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.json")
	original := []byte(`{"model":"opus"}`)
	if err := os.WriteFile(settingsPath, original, 0o600); err != nil {
		t.Fatal(err)
	}

	srv := newTestServer(t)
	srv.SetInstallPaths(settingsPath, filepath.Join(dir, "bin"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/install/statusline", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}

	proposedStr, ok := body["proposed"].(string)
	if !ok || proposedStr == "" {
		t.Fatal("response missing 'proposed' string field")
	}

	var proposed map[string]any
	if err := json.Unmarshal([]byte(proposedStr), &proposed); err != nil {
		t.Fatalf("proposed is not valid JSON: %v", err)
	}
	if proposed["model"] != "opus" {
		t.Error("existing 'model' key must be preserved in proposed")
	}
	if _, ok := proposed["statusLine"]; !ok {
		t.Error("proposed must contain statusLine")
	}

	// Non-mutation guarantee: the on-disk file must be unchanged.
	after, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings after call: %v", err)
	}
	if string(after) != string(original) {
		t.Errorf("settings.json was mutated: got %q, want %q", after, original)
	}
}

// TestHandleProposeStatusline_MissingFile verifies that when the userSettingsPath
// does not exist (fresh install), the handler still returns 200 with a valid
// proposal containing statusLine.
func TestHandleProposeStatusline_MissingFile(t *testing.T) {
	dir := t.TempDir()
	// Point to a file that does not exist.
	nonExistentPath := filepath.Join(dir, "does-not-exist", "settings.json")

	srv := newTestServer(t)
	srv.SetInstallPaths(nonExistentPath, filepath.Join(dir, "bin"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/install/statusline", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}

	proposedStr, ok := body["proposed"].(string)
	if !ok || proposedStr == "" {
		t.Fatal("response missing 'proposed' string field")
	}

	var proposed map[string]any
	if err := json.Unmarshal([]byte(proposedStr), &proposed); err != nil {
		t.Fatalf("proposed is not valid JSON: %v", err)
	}
	if _, ok := proposed["statusLine"]; !ok {
		t.Error("fresh proposal must contain statusLine")
	}
}
