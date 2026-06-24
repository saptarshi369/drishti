// Package api — Module 7 Settings handlers. GET serves the editable config view;
// PUT validates input, saves config.toml (Drishti's OWN file), and reports
// whether a restart is needed. It never writes the user's ~/.claude files.
package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/saptarshi369/drishti/internal/config"
	"github.com/saptarshi369/drishti/internal/services"
	"github.com/saptarshi369/drishti/internal/settings"
)

// handleGetSettings serves the current settings view, including a best-effort
// disk-usage estimate (db_bytes, backup_bytes) and MCP server names from the
// inventory, both appended after BuildView.
func (s *Server) handleGetSettings(w http.ResponseWriter, _ *http.Request) {
	// Snapshot once under a read lock so all field reads are consistent and
	// no concurrent SetConfig call can mutate cfg mid-handler.
	cfg := s.snapshotConfig()
	view := settings.BuildView(cfg, s.version)

	// Resolve backupDir: use the configured value if set; otherwise fall back
	// to the conventional default <data_dir>/backups (spec §6 backup_dir default).
	backupDir := cfg.BackupDir
	if backupDir == "" {
		backupDir = filepath.Join(cfg.DataDir, "backups")
	}

	// DiskEstimate is best-effort: missing paths contribute 0, never errors.
	view.DBBytes, view.BackupBytes = services.DiskEstimate(cfg.DataDir, backupDir)

	// MCP servers come from the inventory the daemon already resolved (no live
	// `claude mcp list` probe yet — deferred). Best-effort: a nil store (test
	// servers) or a query error just yields an empty list; never errors the
	// response (§14 failsafe: degrade, don't die).
	if s.st != nil {
		if rows, err := s.st.ListResolved("mcp", s.currentDefaultRoot(), true); err == nil {
			names := make([]string, 0, len(rows))
			for _, r := range rows {
				names = append(names, r.Name)
			}
			view.MCPServers = names
		}
	}

	writeJSON(w, http.StatusOK, view)
}

// handlePutSettings validates the input, persists config.toml, and returns
// {saved, restart_required, warnings}. Bad input → typed 400 (never a 500).
func (s *Server) handlePutSettings(w http.ResponseWriter, r *http.Request) {
	var in settings.Input
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		apiError(w, http.StatusBadRequest, "settings_invalid", "malformed body", false)
		return
	}
	// Snapshot the current config once; all validation and comparison below
	// must use this consistent snapshot, not repeated live reads.
	cur := s.snapshotConfig()
	newCfg, warns, err := settings.Validate(cur, in)
	if err != nil {
		apiError(w, http.StatusBadRequest, "settings_invalid", err.Error(), false)
		return
	}
	if err := config.Save(newCfg); err != nil {
		apiError(w, http.StatusInternalServerError, "settings_save_failed", "could not write config", true)
		return
	}
	restart := settings.RestartRequired(cur, newCfg)
	// Reflect immediately for subsequent GETs via the locking setter.
	s.SetConfig(newCfg)
	msgs := make([]string, 0, len(warns))
	for _, x := range warns {
		msgs = append(msgs, string(x))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"saved": true, "restart_required": restart, "warnings": msgs,
	})
}
