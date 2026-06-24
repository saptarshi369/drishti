package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/saptarshi369/drishti/internal/config"
)

func TestListDirsRejectsOutsideHome(t *testing.T) {
	home := t.TempDir()
	_, err := listDirsWithin(home, "/etc")
	if err == nil {
		t.Fatal("listing outside home must be refused")
	}
}

func TestListDirsListsSubdirs(t *testing.T) {
	home := t.TempDir()
	os.Mkdir(filepath.Join(home, "proj"), 0o755)
	os.WriteFile(filepath.Join(home, "file.txt"), []byte("x"), 0o644)
	dirs, err := listDirsWithin(home, home)
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) != 1 || filepath.Base(dirs[0]) != "proj" {
		t.Errorf("dirs = %v, want [proj] (dirs only)", dirs)
	}
}

// TestRootsRouteTraversalGuard verifies that GET /api/roots?path=/etc is wired
// into the router and returns 400 (not 404 and not 200).
func TestRootsRouteTraversalGuard(t *testing.T) {
	s := NewServer("test", nil)
	cfg := config.Default()
	cfg.DataDir = t.TempDir()
	s.SetConfig(cfg)
	h := s.Handler()

	r := httptest.NewRequest(http.MethodGet, "/api/roots?path=/etc", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("GET /api/roots?path=/etc: status = %d, want 400 (traversal denied)", w.Code)
	}
}
