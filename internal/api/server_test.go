package api

import (
	"net/http"
	"net/http/httptest"
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
