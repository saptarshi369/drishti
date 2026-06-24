// Package api — Module 7 folder picker. Lists directories so the browser (which
// cannot see the filesystem) can pick watched folders. The lister is constrained
// to the user's home tree so this localhost endpoint never enumerates the whole
// filesystem.
package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/saptarshi369/drishti/internal/config"
)

// listDirsWithin returns the immediate subdirectories of path, but only if path
// resolves inside home. Returns os.ErrPermission for any path outside home.
// This is the security-critical part: it uses filepath.Rel to detect traversal
// attempts like "/etc", "..", or "/home/../etc" after filepath.Clean normalises
// the input.
func listDirsWithin(home, path string) ([]string, error) {
	if path == "" {
		path = home
	}
	// filepath.Clean normalises away ".." components and symlink-independent
	// double slashes, so "/home/../etc" becomes "/etc" before the guard runs.
	clean := filepath.Clean(path)

	// Reject paths that escape home. filepath.Rel returns the path of clean
	// relative to home. If the result starts with ".." the path is outside.
	rel, err := filepath.Rel(home, clean)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return nil, os.ErrPermission
	}

	entries, err := os.ReadDir(clean)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(clean, e.Name()))
		}
	}
	return dirs, nil
}

// handleListDirs serves GET /api/roots?path= — the folder picker's directory feed.
// It resolves the user's home directory, then delegates to listDirsWithin for the
// security-guarded directory listing. Any path outside ~ returns 400.
func (s *Server) handleListDirs(w http.ResponseWriter, r *http.Request) {
	home, err := os.UserHomeDir()
	if err != nil {
		apiError(w, http.StatusInternalServerError, "roots_no_home", "cannot resolve home", true)
		return
	}
	dirs, err := listDirsWithin(home, r.URL.Query().Get("path"))
	if err != nil {
		apiError(w, http.StatusBadRequest, "roots_denied", "path not allowed", false)
		return
	}
	// Force non-nil slice so JSON encodes [] not null when there are no subdirs.
	if dirs == nil {
		dirs = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"home": home, "dirs": dirs})
}

// handleSetRoots persists the watched-folder list to config.toml [roots].
// It accepts {"paths": ["..."]} and writes atomically via config.Save.
// On success it immediately updates the in-memory config so subsequent GETs
// reflect the change without waiting for the daemon's ~10s reload cycle.
func (s *Server) handleSetRoots(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Paths []string `json:"paths"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiError(w, http.StatusBadRequest, "roots_invalid", "malformed body", false)
		return
	}
	// Snapshot the current config, then mutate a copy so s.cfg is only updated
	// after a successful save. Using snapshotConfig() prevents a torn read of
	// the []string Roots field if a concurrent SetConfig is in flight.
	newCfg := s.snapshotConfig()
	newCfg.Roots = body.Paths
	if err := config.Save(newCfg); err != nil {
		apiError(w, http.StatusInternalServerError, "roots_save_failed", "could not write config", true)
		return
	}
	// Commit via the locking setter so the live snapshot is updated atomically
	// and GET /api/settings sees the change immediately without a scheduler tick.
	s.SetConfig(newCfg)
	writeJSON(w, http.StatusOK, map[string]any{"saved": true})
}
