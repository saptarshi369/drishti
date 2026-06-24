// Package api — Module 7 install handler.
//
// This file implements the non-mutating settings.json proposal endpoint.
// Drishti NEVER writes the user's ~/.claude/settings.json. The handler reads
// the existing file, passes it to install.ProposeSettingsJSON, and returns the
// merged proposal for the user to copy and apply themselves.
package api

import (
	"net/http"
	"os"

	"github.com/saptarshi369/drishti/internal/install"
)

// handleProposeStatusline serves GET /api/install/statusline. It reads the
// user's existing settings.json (missing file → nil, treated as {} by the
// generator) and returns a proposed document with Drishti's statusLine entry
// merged in. The on-disk file is never written — non-mutation is guaranteed
// by the install package's pure-function design.
//
// Response 200: { "proposed": "<pretty JSON string>", "added": ["statusLine"],
// "path": "<absolute path to the user's settings.json>" }
// Response 400: typed error "install_bad_existing" when the existing file is
// present but contains invalid JSON.
func (s *Server) handleProposeStatusline(w http.ResponseWriter, _ *http.Request) {
	// os.ReadFile returns (nil, err) when the file is absent. We ignore the
	// error intentionally: a missing file is the "fresh install" path and is
	// treated as {} by ProposeSettingsJSON (no mutation, no panic).
	existing, _ := os.ReadFile(s.userSettingsPath)

	proposed, added, err := install.ProposeSettingsJSON(existing, s.helperBinDir)
	if err != nil {
		// The existing file is present but not valid JSON. Return a typed 400
		// so the UI can tell the user to fix their settings.json first.
		apiError(w, http.StatusBadRequest, "install_bad_existing", err.Error(), false)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"proposed": string(proposed),
		"added":    added,
		"path":     s.userSettingsPath,
	})
}
