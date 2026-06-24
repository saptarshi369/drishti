package api

import (
	"io/fs"
	"net/http"
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

// currentDefaultRoot returns the current default root under a read lock.
func (s *Server) currentDefaultRoot() string {
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
	mux.Handle("/", http.FileServer(http.FS(sub)))
	return mux
}
