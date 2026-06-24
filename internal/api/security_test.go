package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/store"
)

// TestHandleSecurity_OK seeds one finding and verifies the handler returns a
// 200 with a snapshot reflecting that finding.
func TestHandleSecurity_OK(t *testing.T) {
	// Open a temporary in-memory-style store for this test.
	st, err := store.Open(t.TempDir() + "/t.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	// Seed one high-severity finding for the default root ("").
	if err := st.ReplaceSecurityFindings("claude", "", []model.Finding{
		{RuleID: "bypass-permissions-mode", Severity: "high", Title: "t", TargetKey: "user:settings.json", Detail: "d", Remediation: "r", Scope: "user"},
	}); err != nil {
		t.Fatal(err)
	}

	srv := NewServer("test", st)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/security", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	// Decode and assert the snapshot fields.
	var snap model.SecuritySnapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snap); err != nil {
		t.Fatal(err)
	}
	if snap.Total != 1 || snap.Counts["high"] != 1 {
		t.Fatalf("snapshot = %+v", snap)
	}
}

// TestHandleSecurity_UnknownAgent verifies that an unrecognised ?agent= value
// returns a typed 400 without touching the store.
func TestHandleSecurity_UnknownAgent(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/t.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	srv := NewServer("test", st)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/security?agent=codex", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
