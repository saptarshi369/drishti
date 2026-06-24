// Package app supervises the whole daemon: the startup reconcile ladder, the
// scheduler, the HTTP server, and graceful shutdown. It is the only place that
// wires the other packages together; nothing imports app.
package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/saptarshi369/drishti/internal/api"
	"github.com/saptarshi369/drishti/internal/config"
	"github.com/saptarshi369/drishti/internal/ingest"
	"github.com/saptarshi369/drishti/internal/install"
	"github.com/saptarshi369/drishti/internal/logging"
	"github.com/saptarshi369/drishti/internal/security"
	"github.com/saptarshi369/drishti/internal/services"
	"github.com/saptarshi369/drishti/internal/skills"
	claude "github.com/saptarshi369/drishti/internal/sources/claude"
	"github.com/saptarshi369/drishti/internal/store"
)

// OpenStoreWithQuarantine opens the DB, runs integrity_check, and — if the file
// is corrupt — moves it aside (kept for debugging) and recreates a fresh one.
// Because the DB is a rebuildable cache, no user data is lost. Returns whether
// a rebuild happened.
func OpenStoreWithQuarantine(dbPath string, lg *slog.Logger) (*store.Store, bool, error) {
	if lg == nil {
		lg = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	st, err := store.Open(dbPath)
	if err == nil {
		ok, ierr := st.IntegrityOK()
		if ierr == nil && ok {
			return st, false, nil
		}
		_ = st.Close()
	}
	// Quarantine and rebuild.
	if _, statErr := os.Stat(dbPath); statErr == nil {
		side := fmt.Sprintf("%s.corrupt.%d", dbPath, time.Now().UnixMilli())
		if mvErr := os.Rename(dbPath, side); mvErr != nil {
			return nil, false, fmt.Errorf("quarantine rename: %w", mvErr)
		}
		lg.Warn("quarantined corrupt db; rebuilding from source", "sidecar", side)
		// Remove WAL/SHM siblings so the fresh DB starts clean.
		_ = os.Remove(dbPath + "-wal")
		_ = os.Remove(dbPath + "-shm")
	}
	fresh, err := store.Open(dbPath)
	if err != nil {
		return nil, false, fmt.Errorf("recreate db: %w", err)
	}
	return fresh, true, nil
}

// acquireLock takes an exclusive advisory lock so two daemons don't share a DB.
// It uses O_CREATE|O_EXCL on a lock file; release removes it.
func acquireLock(dataDir string) (func(), error) {
	lockPath := filepath.Join(dataDir, "drishti.lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("another drishti instance appears to be running (%s): %w", lockPath, err)
	}
	_, _ = fmt.Fprintf(f, "%d\n", os.Getpid())
	_ = f.Close()
	return func() { _ = os.Remove(lockPath) }, nil
}

// Run is the daemon entrypoint. It returns nil on clean shutdown.
func Run(ctx context.Context, version string) error {
	dataDir := resolveDataDir()
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	cfg, warns, _ := config.Load(dataDir)
	lg, err := logging.New(dataDir, cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("init logging: %w", err)
	}
	for _, wmsg := range warns {
		lg.Warn("config", "note", string(wmsg))
	}

	// security-rules.toml lives next to the DB; write the documented default on
	// first run (never clobbering user edits) so the file is there to edit.
	rulesPath := filepath.Join(dataDir, "security-rules.toml")
	if err := security.EnsureRulesFile(rulesPath); err != nil {
		lg.Warn("could not write default security rules", "path", rulesPath, "err", err)
	}

	// skills-analytics.toml lives next to the DB; write the documented default
	// on first run (never clobbering user edits) so the file is there to edit.
	thresholdsPath := filepath.Join(dataDir, "skills-analytics.toml")
	if err := skills.EnsureThresholdsFile(thresholdsPath); err != nil {
		lg.Warn("could not write default skills thresholds", "path", thresholdsPath, "err", err)
	}

	release, err := acquireLock(dataDir)
	if err != nil {
		return err
	}
	defer release()

	st, rebuilt, err := OpenStoreWithQuarantine(filepath.Join(dataDir, "drishti.db"), lg)
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()
	if rebuilt {
		lg.Warn("rebuilt local cache from source (no data lost)")
	}

	// Inject the pricing fn so the store stamps est_cost_usd as it folds each
	// usage_rollup row AT INGEST. This keeps the hot read/broadcast path read-only:
	// OverviewSnapshot/UsageSnapshot no longer rewrite the whole table under the
	// write lock on every 1s tick (the Overview-slowness fix). Then backfill once
	// to price any rows ingested before this wiring existed; a failure is logged and
	// ignored so startup never dies over a cost recompute (§14).
	st.SetCostFn(services.Cost)
	if err := st.BackfillRollupCost(services.Cost); err != nil {
		lg.Warn("startup cost backfill", "err", err)
	}

	// Startup reconcile ladder: cold/incremental scan of all transcripts.
	roots := []string{filepath.Join(userHome(), ".claude", "projects")}
	rec := ingest.New(st, roots, lg)
	if err := rec.ScanAll(); err != nil {
		lg.Warn("initial scan error", "err", err)
	}

	// Inventory: discover and resolve skills/agents/hooks/MCP across the user
	// global location (~/.claude) and every configured project root. We do this
	// once at startup so the UI has data immediately; the scheduler re-runs it
	// every ~10 s so newly installed skills appear without a daemon restart.
	invLocs := inventoryLocations(cfg)
	if err := services.RefreshInventory(st, invLocs, security.LoadRulesFromPath(rulesPath, lg)); err != nil {
		// Degrade gracefully (spec §14): warn and continue. Inventory data will
		// be stale until the scheduler picks it back up.
		lg.Warn("initial inventory refresh", "err", err)
	}

	// helperBinDir is Drishti's own script directory (~/.drishti/bin/). Write the
	// statusline helper into it now so the path referenced in the proposed
	// settings.json always resolves when the user applies it. Failure is logged
	// and ignored — startup must not fail because of a script write error (§14).
	helperBinDir := filepath.Join(dataDir, "bin")
	if err := install.EnsureHelperScripts(helperBinDir); err != nil {
		lg.Warn("could not write helper scripts", "dir", helperBinDir, "err", err)
	}

	srv := api.NewServer(version, st)
	// Default the inventory view to the primary project root (pwd by default) so
	// the UI shows the project's merged user+project config without a ?root= param.
	srv.SetDefaultRoot(primaryRoot(cfg))
	srv.SetContextWindowTokens(cfg.ContextWindowTokens)
	// Load the over-triggering thresholds for the Skills screen. The scheduler
	// reloads them every ~10 s (below) so edits are picked up without a restart.
	srv.SetSkillThresholds(skills.LoadThresholdsFromPath(thresholdsPath, lg))
	// Seed the server with the current config so the Settings screen can read + edit it.
	srv.SetConfig(cfg)
	// Give the server the paths to the two user-editable rule files so that
	// PUT /api/thresholds and PUT /api/rules know where to write. The ~10 s
	// scheduler reload (above) picks up the new files automatically.
	srv.SetConfigFilePaths(rulesPath, thresholdsPath)
	// Give the install handler the path to the user's settings.json (read-only)
	// and to Drishti's own bin dir (for script path generation). The handler
	// NEVER writes the user's file — non-mutation is guaranteed.
	userSettingsPath := filepath.Join(userHome(), ".claude", "settings.json")
	srv.SetInstallPaths(userSettingsPath, helperBinDir)
	httpSrv := &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.BindAddr, cfg.Port), Handler: srv.Handler()}

	// Watcher: on new data, push a fresh snapshot to all SSE clients.
	safego("watcher", lg, func() {
		rec.Watch(ctx, func(inserted int) {
			if inserted > 0 {
				srv.BroadcastSnapshot()
			}
		})
	})

	// Scheduler: periodic snapshot heartbeat keeps gauges fresh (coalesced ≤1/s).
	// Every 10th tick (~10 s) it also re-scans the inventory so newly installed
	// skills/agents appear without a daemon restart. Failures are logged and
	// isolated so they never stop the heartbeat loop.
	safego("scheduler", lg, func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		// invTick counts scheduler ticks; inventory is re-scanned every 10 ticks.
		invTick := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				srv.BroadcastSnapshot()
				invTick++
				if invTick%10 == 0 {
					// Reload config.toml so settings saved via PUT /api/settings
					// hot-apply within ~10s (theme/accent, retention, roots,
					// window_tokens). port/bind cannot rebind a live listener, so
					// those need a restart (the PUT response flags that). A reload
					// failure is logged + ignored so the heartbeat never stops (§14).
					if reloaded, _, lerr := config.Load(cfg.DataDir); lerr == nil {
						cfg = reloaded
						srv.SetConfig(cfg)
						srv.SetDefaultRoot(primaryRoot(cfg))
						srv.SetContextWindowTokens(cfg.ContextWindowTokens)
						invLocs = inventoryLocations(cfg) // re-derive watched roots
					} else {
						lg.Warn("config reload", "err", lerr)
					}
					// Reloading the rules file on every inventory refresh means
					// edits to security-rules.toml are picked up within ~10 s —
					// no separate file watcher needed.
					if err := services.RefreshInventory(st, invLocs, security.LoadRulesFromPath(rulesPath, lg)); err != nil {
						lg.Warn("inventory rescan", "err", err)
					} else {
						// Notify connected UI clients so they can refetch /api/inventory.
						// Payload is nil for Module 1; the client triggers a full fetch.
						srv.Hub().Broadcast(api.Message{
							Type:    "inventory_changed",
							TS:      time.Now().UnixMilli(),
							Payload: nil,
						})
					}
					// Pick up edits to skills-analytics.toml within ~10 s, same
					// cadence as the security rules reload above.
					srv.SetSkillThresholds(skills.LoadThresholdsFromPath(thresholdsPath, lg))
				}
			}
		}
	})

	errCh := make(chan error, 1)
	go func() { errCh <- httpSrv.ListenAndServe() }()
	lg.Info("listening", "addr", httpSrv.Addr, "version", version)

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		lg.Info("shutting down")
		return shutdown(httpSrv, lg)
	}
}

// shutdown drains HTTP, then lets deferred Close() checkpoint + close the DB.
func shutdown(httpSrv *http.Server, lg *slog.Logger) error {
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutCtx); err != nil {
		lg.Warn("http shutdown", "err", err)
		return err
	}
	return nil
}

// resolveDataDir honors DRISHTI_DATA_DIR, defaulting to ~/.drishti.
func resolveDataDir() string {
	if d := os.Getenv("DRISHTI_DATA_DIR"); d != "" {
		return d
	}
	return filepath.Join(userHome(), ".drishti")
}

// userHome returns the user's home dir, falling back to "." if unknown.
func userHome() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return h
}

// inventoryLocations builds the set of discovery scopes for the inventory
// pipeline. It always includes one user-global location (reading ~/.claude and
// ~/.claude.json). In addition, it appends one claude.Locations entry per
// configured project root in cfg.Roots so project-scoped skills/agents are also
// resolved.
//
// When no roots are configured, it defaults to the user's home directory as
// the single project root. This broad default ensures the daemon discovers all
// projects under ~ on first launch; users can narrow the scope by adding
// explicit roots in their config.
//
// The returned slice is passed to services.RefreshInventory. The user-global
// entry carries no ProjectRoot so the store keys it as the user scope.
func inventoryLocations(cfg config.Config) []claude.Locations {
	home := userHome()
	// User-global location: reads ~/.claude/ (skills/agents/hooks) and
	// ~/.claude.json (MCP user/local settings).
	locs := []claude.Locations{{
		UserClaudeDir:  filepath.Join(home, ".claude"),
		UserClaudeJSON: filepath.Join(home, ".claude.json"),
		// ProjectRoot is intentionally empty: this entry covers the user scope only.
	}}
	// One extra entry per project root. Each project inherits the user-global
	// paths so user-scope items also feed into per-project resolution.
	for _, root := range effectiveRoots(cfg) {
		locs = append(locs, claude.Locations{
			UserClaudeDir:  filepath.Join(home, ".claude"),
			UserClaudeJSON: filepath.Join(home, ".claude.json"),
			ProjectRoot:    root,
		})
	}
	return locs
}

// effectiveRoots returns the project roots to scan: the configured cfg.Roots,
// or the user's home directory when none are configured. (Module 7 changed this
// default from the working directory to ~: pwd is meaningless once Drishti is
// installed permanently, so the user adds watched folders explicitly instead.)
func effectiveRoots(cfg config.Config) []string {
	if len(cfg.Roots) > 0 {
		return cfg.Roots
	}
	return []string{userHome()}
}

// primaryRoot is the project root the inventory UI defaults to (the first
// effective root, i.e. the user's home dir when nothing is configured). "" when there is none.
func primaryRoot(cfg config.Config) string {
	roots := effectiveRoots(cfg)
	if len(roots) == 0 {
		return ""
	}
	return roots[0]
}
