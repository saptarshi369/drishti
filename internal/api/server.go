package api

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/saptarshi369/drishti/internal/config"
	"github.com/saptarshi369/drishti/internal/skills"
	"github.com/saptarshi369/drishti/internal/store"
	"github.com/saptarshi369/drishti/web"
)

// Server wires the HTTP routes and holds shared dependencies.
type Server struct {
	version string
	hub     *Hub
	st      *store.Store

	// mu guards the four fields below that are written by the daemon's ~10s
	// scheduler goroutine (via the Set* setters) AND read by HTTP handler
	// goroutines. Without a mutex, concurrent access causes a data race
	// detected by -race. The set-once-before-serve path fields (rulesPath,
	// thresholdsPath, userSettingsPath, helperBinDir) are NOT guarded here
	// because they are written once before ListenAndServe and never mutated.
	mu sync.RWMutex

	// defaultRoot is the project root used for inventory queries that omit a
	// ?root= param. It is the daemon's primary root (~ by default); that
	// resolved set already merges user + project scope. Empty means user-global.
	// Guarded by mu.
	defaultRoot string
	// selectedRoot is the root the user picked in the top-bar selector (PUT
	// /api/active-root). When non-empty it OVERRIDES defaultRoot app-wide so every
	// screen (incl. the SSE-driven Overview) re-scopes to it. It is in-memory only:
	// a daemon restart resets to the configured primary. Guarded by mu.
	selectedRoot string
	// contextWindowTokens is the denominator for the Context-Budget tax %; set
	// from config by the daemon. 0 means "unset" → the handler treats pct as 0.
	// Guarded by mu.
	contextWindowTokens int
	// skillThresholds tunes the over-triggering flag on the Skills screen. The
	// daemon sets it at startup and refreshes it each scheduler cycle so edits
	// to skills-analytics.toml are picked up live. Zero value flags nothing.
	// Guarded by mu.
	skillThresholds skills.Thresholds
	// cfg is the live config snapshot, used by the settings handlers + served at
	// GET /api/settings. The daemon refreshes it on each scheduler reload.
	// Guarded by mu.
	cfg config.Config

	// rulesPath is the absolute path to security-rules.toml. Set by the daemon
	// via SetConfigFilePaths so PUT /api/rules knows where to write.
	rulesPath string
	// thresholdsPath is the absolute path to skills-analytics.toml. Set by the
	// daemon via SetConfigFilePaths so PUT /api/thresholds knows where to write.
	thresholdsPath string
	// userSettingsPath is the absolute path to the user's ~/.claude/settings.json.
	// The daemon sets this via SetInstallPaths. The install handler reads it to
	// produce a proposal but NEVER writes it — non-mutation is guaranteed.
	userSettingsPath string
	// helperBinDir is the absolute path to ~/.drishti/bin/ where Drishti's own
	// helper scripts are installed. The daemon sets this via SetInstallPaths.
	helperBinDir string
}

// NewServer constructs a Server bound to a store. version is build-stamped.
func NewServer(version string, st *store.Store) *Server {
	return &Server{version: version, hub: NewHub(), st: st}
}

// SetDefaultRoot sets the project root used for inventory queries that omit a
// ?root= param. The daemon calls this with its primary root so the UI shows the
// project (home directory ~) view by default.
func (s *Server) SetDefaultRoot(root string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.defaultRoot = root
}

// SetContextWindowTokens sets the context-window denominator for the
// Context-Budget percentage (from config.ContextWindowTokens).
func (s *Server) SetContextWindowTokens(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.contextWindowTokens = n
}

// SetSkillThresholds sets the over-triggering thresholds for the Skills screen.
// The daemon calls this at startup and on each ~10 s scheduler tick so edits to
// skills-analytics.toml take effect without a restart.
func (s *Server) SetSkillThresholds(t skills.Thresholds) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.skillThresholds = t
}

// SetConfig stores the current config snapshot for the settings handlers.
// The daemon calls this on each ~10 s scheduler tick so the settings screen
// always reflects the current on-disk state.
func (s *Server) SetConfig(cfg config.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg
}

// snapshotConfig returns a copy of the live config under a read lock.
// Callers must use the returned copy and not hold references into s.cfg.
func (s *Server) snapshotConfig() config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

// SetSelectedRoot records the user's top-bar root choice. Empty clears it (back to
// the daemon's primary root). Callers should pass a validated root (one of
// rootOptions); validation lives in the PUT handler, not here.
func (s *Server) SetSelectedRoot(root string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selectedRoot = root
}

// currentDefaultRoot returns the active view root under a read lock. It is the
// user's top-bar selection; the default (and the explicit "All" choice) is the
// empty string, meaning user-global inventory + no usage/event filter. This single
// accessor is what every screen scopes to, so a selection re-scopes them all.
func (s *Server) currentDefaultRoot() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectedRoot
}

// currentPrimaryRoot returns the daemon's configured primary root (ignoring any
// top-bar selection), so the selector UI can show which option is the default.
func (s *Server) currentPrimaryRoot() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.defaultRoot
}

// currentContextWindowTokens returns the current context-window token count under a read lock.
func (s *Server) currentContextWindowTokens() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.contextWindowTokens
}

// currentSkillThresholds returns the current skill thresholds under a read lock.
func (s *Server) currentSkillThresholds() skills.Thresholds {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.skillThresholds
}

// SetConfigFilePaths records the on-disk paths for the two user-editable rule
// files. The daemon calls this once after resolving the data directory so that
// PUT /api/rules and PUT /api/thresholds know where to write. The ~10s
// scheduler reload picks up the new files automatically.
func (s *Server) SetConfigFilePaths(rulesPath, thresholdsPath string) {
	s.rulesPath = rulesPath
	s.thresholdsPath = thresholdsPath
}

// SetInstallPaths records the path to the user's ~/.claude/settings.json (for
// reading only — Drishti never writes it) and the path to ~/.drishti/bin/
// (Drishti's own helper-script directory). The daemon calls this once at
// startup so the install handler has the correct paths.
func (s *Server) SetInstallPaths(userSettingsPath, helperBinDir string) {
	s.userSettingsPath = userSettingsPath
	s.helperBinDir = helperBinDir
}

// Hub exposes the SSE hub so the daemon can broadcast counters/status to it.
func (s *Server) Hub() *Hub { return s.hub }

// Handler returns the root http.Handler: the embedded UI plus the API routes.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/overview", s.handleOverview)
	mux.HandleFunc("GET /api/update/status", s.handleUpdateStatus)
	mux.HandleFunc("GET /api/stream", s.streamHandler)
	mux.HandleFunc("GET /api/inventory", s.handleInventory)
	mux.HandleFunc("GET /api/inventory/{id}/why", s.handleInventoryWhy)
	mux.HandleFunc("GET /api/activity", s.handleActivity)
	mux.HandleFunc("GET /api/activity/events", s.handleActivityEvents)
	mux.HandleFunc("GET /api/usage", s.handleUsage)
	mux.HandleFunc("GET /api/quota", s.handleQuota)
	mux.HandleFunc("POST /api/quota/sample", s.handleQuotaSample)
	mux.HandleFunc("GET /api/context-budget", s.handleContextBudget)
	mux.HandleFunc("GET /api/security", s.handleSecurity)
	mux.HandleFunc("GET /api/skills", s.handleSkills)
	mux.HandleFunc("GET /api/settings", s.handleGetSettings)
	mux.HandleFunc("PUT /api/settings", s.handlePutSettings)
	mux.HandleFunc("GET /api/roots", s.handleListDirs)
	mux.HandleFunc("PUT /api/roots", s.handleSetRoots)
	mux.HandleFunc("GET /api/active-root", s.handleGetActiveRoot)
	mux.HandleFunc("PUT /api/active-root", s.handleSetActiveRoot)
	mux.HandleFunc("GET /api/thresholds", s.handleGetThresholds)
	mux.HandleFunc("PUT /api/thresholds", s.handlePutThresholds)
	mux.HandleFunc("PUT /api/rules", s.handlePutRules)
	mux.HandleFunc("GET /api/install/statusline", s.handleProposeStatusline)

	// Serve the embedded SvelteKit bundle. fs.Sub strips the build/ prefix so
	// "/" maps to index.html.
	sub, err := fs.Sub(web.FS, "build")
	if err != nil {
		panic(err) // compile-time embed guarantees this never fails at runtime
	}
	mux.Handle("/", spaFileServer(sub))
	return mux
}

// spaFileServer serves the embedded single-page-app bundle with a client-route
// fallback. SvelteKit (adapter-static, fallback:index.html) emits exactly one
// HTML shell; every page is reached by client-side routing. A plain
// http.FileServer therefore 404s on a hard navigation / refresh / bookmark of a
// sub-route like /inventory, because no such file exists. This wrapper serves the
// requested file when it exists (index.html, /_app assets, …) and otherwise
// rewrites the request to "/" so the SPA shell is returned with 200 — letting the
// browser hydrate and route to the intended page. Genuinely-missing files keep
// 404ing only at the asset level: a navigation-looking path (no real file) gets
// the shell, which is the standard SPA-server contract.
func spaFileServer(sub fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Map the URL path to an FS path: strip the leading slash and clean it.
		// path.Clean keeps things like "../" from escaping the embedded root.
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "index.html" // "/" → the shell
		}
		// If a real, non-directory file backs this path, serve it untouched so
		// assets keep their correct content-type.
		if info, err := fs.Stat(sub, name); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		// The path has no backing file. If it looks like a static asset (it has a
		// file extension, e.g. /_app/…​.js), let FileServer return its normal 404 —
		// masking a broken bundle reference with HTML would hide real errors.
		if path.Ext(name) != "" {
			fileServer.ServeHTTP(w, r) // 404 for the genuinely-missing asset
			return
		}
		// Otherwise it's a client route (e.g. /inventory): return the SPA shell so
		// the browser can hydrate and route. Rewrite a CLONE to "/" so FileServer
		// emits index.html (200, text/html) without mutating the caller's request.
		shell := r.Clone(r.Context())
		shell.URL.Path = "/"
		fileServer.ServeHTTP(w, shell)
	})
}
