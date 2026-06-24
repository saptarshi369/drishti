package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/saptarshi369/drishti/internal/config"
)

// TestActiveRootSelectChangesScope verifies the top-bar scope selector: the default
// scope is "All" (empty string); PUT /api/active-root switches it to a folder (which
// scopes the Overview/inventory/etc.) or back to All (""); GET lists the configured
// folders + the current selection. An unknown root is rejected and leaves scope intact.
func TestActiveRootSelectChangesScope(t *testing.T) {
	srv := NewServer("test", nil)
	cfg := config.Default()
	cfg.Roots = []string{"/tmp/proj-a", "/tmp/proj-b"}
	srv.SetConfig(cfg)
	h := srv.Handler()

	// Default scope is "All" (empty), not any configured folder.
	if got := srv.currentDefaultRoot(); got != "" {
		t.Fatalf("initial scope = %q, want \"\" (All)", got)
	}

	// GET lists the configured folders and the current selection (All = "").
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/active-root", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", rec.Code)
	}
	var got struct {
		Current string   `json:"current"`
		Roots   []string `json:"roots"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Current != "" {
		t.Errorf("GET current = %q, want \"\" (All)", got.Current)
	}
	if !contains(got.Roots, "/tmp/proj-a") || !contains(got.Roots, "/tmp/proj-b") {
		t.Errorf("GET roots = %v, want both configured folders", got.Roots)
	}

	// PUT switches the active scope to proj-b.
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodPut, "/api/active-root",
		strings.NewReader(`{"root":"/tmp/proj-b"}`)))
	if rec2.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200", rec2.Code)
	}
	if got := srv.currentDefaultRoot(); got != "/tmp/proj-b" {
		t.Errorf("after PUT scope = %q, want /tmp/proj-b", got)
	}

	// PUT "" selects All again (no filter).
	recAll := httptest.NewRecorder()
	h.ServeHTTP(recAll, httptest.NewRequest(http.MethodPut, "/api/active-root",
		strings.NewReader(`{"root":""}`)))
	if recAll.Code != http.StatusOK {
		t.Fatalf("PUT All status = %d, want 200", recAll.Code)
	}
	if got := srv.currentDefaultRoot(); got != "" {
		t.Errorf("after PUT All scope = %q, want \"\"", got)
	}

	// Switch to proj-b, then an unknown root must be rejected and not change scope.
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPut, "/api/active-root",
		strings.NewReader(`{"root":"/tmp/proj-b"}`)))
	rec3 := httptest.NewRecorder()
	h.ServeHTTP(rec3, httptest.NewRequest(http.MethodPut, "/api/active-root",
		strings.NewReader(`{"root":"/tmp/not-a-root"}`)))
	if rec3.Code != http.StatusBadRequest {
		t.Errorf("PUT unknown status = %d, want 400", rec3.Code)
	}
	if got := srv.currentDefaultRoot(); got != "/tmp/proj-b" {
		t.Errorf("rejected PUT must not change scope, got %q", got)
	}
}

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}
