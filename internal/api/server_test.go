package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/saptarshi369/drishti/internal/config"
	"github.com/saptarshi369/drishti/internal/skills"
)

func TestHandlerServesEmbeddedUI(t *testing.T) {
	srv := NewServer("test", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	if rec.Body.Len() == 0 {
		t.Errorf("GET / returned empty body; expected embedded index.html")
	}
}

// TestHandlerSPAFallback verifies that a hard navigation / refresh to a
// client-side route (e.g. /inventory) serves the SPA shell with 200 instead of
// the bare http.FileServer 404. Without the fallback, refreshing or bookmarking
// any sub-page shows "404 page not found". Genuine asset paths must still 404 so
// a broken bundle reference is not masked by HTML.
func TestHandlerSPAFallback(t *testing.T) {
	srv := NewServer("test", nil)
	h := srv.Handler()

	// A client route that has no corresponding file → SPA shell, 200.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/inventory", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /inventory status = %d, want 200 (SPA fallback)", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("GET /inventory content-type = %q, want text/html", ct)
	}

	// A missing asset under the immutable bundle dir must still 404 — we only
	// fall back for navigations, not for genuinely-absent files.
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/_app/does-not-exist.js", nil))
	if rec2.Code != http.StatusNotFound {
		t.Errorf("GET missing asset status = %d, want 404", rec2.Code)
	}
}

// TestServerConcurrentConfigAccess exercises concurrent setter + reader access so
// the race detector trips when the fields are not protected by a mutex. Run with
// `make test` which pipes through -race; a DATA RACE report here means the guard
// is missing (red), and a clean run means the mutex is working (green).
func TestServerConcurrentConfigAccess(t *testing.T) {
	s := NewServer("test", nil)
	s.SetConfig(config.Default())
	var wg sync.WaitGroup
	wg.Add(2)
	// Writer goroutine: hammers all four scheduler-mutated fields 1000 times.
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			s.SetConfig(config.Default())
			s.SetDefaultRoot("/x")
			s.SetContextWindowTokens(i)
			s.SetSkillThresholds(skills.Thresholds{})
		}
	}()
	// Reader goroutine: reads all four fields 1000 times via the private accessors.
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = s.snapshotConfig()
			_ = s.currentDefaultRoot()
			_ = s.currentContextWindowTokens()
			_ = s.currentSkillThresholds()
		}
	}()
	wg.Wait()
}
