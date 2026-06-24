package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/saptarshi369/drishti/internal/config"
)

// TestActiveRootSelectChangesScope verifies the top-bar root selector: PUT
// /api/active-root switches the global view root (what currentDefaultRoot returns,
// which scopes the Overview/inventory/etc.), and GET lists the selectable roots
// plus the current selection. An unknown root is rejected and leaves scope intact.
func TestActiveRootSelectChangesScope(t *testing.T) {
	srv := NewServer("test", nil)
	cfg := config.Default()
	cfg.Roots = []string{"/tmp/proj-a", "/tmp/proj-b"}
	srv.SetConfig(cfg)
	srv.SetDefaultRoot("/tmp/proj-a") // daemon's primary root
	h := srv.Handler()

	if got := srv.currentDefaultRoot(); got != "/tmp/proj-a" {
		t.Fatalf("initial scope = %q, want /tmp/proj-a", got)
	}

	// GET lists the options (home + both roots) and the current selection.
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
	if got.Current != "/tmp/proj-a" {
		t.Errorf("GET current = %q, want /tmp/proj-a", got.Current)
	}
	if !contains(got.Roots, "/tmp/proj-a") || !contains(got.Roots, "/tmp/proj-b") {
		t.Errorf("GET roots = %v, want both configured roots", got.Roots)
	}

	// PUT switches the active root to proj-b.
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodPut, "/api/active-root",
		strings.NewReader(`{"root":"/tmp/proj-b"}`)))
	if rec2.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200", rec2.Code)
	}
	if got := srv.currentDefaultRoot(); got != "/tmp/proj-b" {
		t.Errorf("after PUT scope = %q, want /tmp/proj-b", got)
	}

	// PUT with a root that is neither home nor a configured root is rejected.
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
