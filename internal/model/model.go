// Package model holds Drishti's plain data structs. It knows nothing about
// SQL or HTTP — every other package depends on it, and it depends on none of
// ours. Keeping these types dependency-free is what lets parsers, the store,
// and the API agree on shapes without coupling.
package model

// SourceFile is one row of the ingestion ledger: how far we have cleanly read
// a given agent file, plus the identity signals used to detect rotation.
type SourceFile struct {
	ID         int64
	AgentCode  string // "claude"
	Kind       string // "transcript"
	AbsPath    string
	FileID     string // platform.FileIdentity result; "" if unavailable
	Size       int64
	MtimeMs    int64
	HeadHash   string
	LastOffset int64
	LastLine   int64
	State      string // active|rotated|missing|error
	ErrorCount int
}

// Event is one ingested fact: prompt | tool_use | skill | blocked | error |
// session_start | session_end. It carries NO raw text (privacy default D8):
// ToolName/SkillName are component NAMES; Status is "error"|"blocked"|"".
type Event struct {
	AgentCode  string
	TypeCode   string // prompt|tool_use|skill|session_start|session_end
	SourceCode string // transcript
	SessionID  string
	TsMs       int64
	ToolName   string // set for tool_use
	SkillName  string // set for skill
	Status     string // "error"|"blocked" for those types; "" otherwise
	DedupeKey  string // agent|session_id|sha1(line) for prompts; agent|session|block.id for tools
}

// SessionDelta is the per-session contribution parsed from one file region:
// token totals to add and the prompt count seen. Day is the local yyyymmdd the
// activity belongs to (used to key usage_rollup).
type SessionDelta struct {
	SessionID    string
	Model        string
	Day          int // yyyymmdd, local
	StartedMs    int64
	InputTokens  int64
	OutputTokens int64
	CacheTokens  int64
	PromptCount  int
}

// Session is the persisted per-conversation spine with cached counters.
type Session struct {
	AgentCode    string
	SessionID    string
	Model        string
	StartedMs    int64
	PromptCount  int
	InputTokens  int64
	OutputTokens int64
	CacheTokens  int64
	EstCostUSD   float64
}

// ParseResult is everything one parser pass extracts from a file region, plus
// how many bytes/lines were COMPLETELY consumed (so ingest advances the ledger
// offset only past whole lines).
type ParseResult struct {
	Events        []Event
	Deltas        []SessionDelta
	BytesConsumed int64
	LinesConsumed int64
	ErrorCount    int
}

// OverviewKPIs is the headline card data for the Overview screen.
type OverviewKPIs struct {
	PromptsToday  int     `json:"prompts_today"`
	SpendTodayUSD float64 `json:"spend_today_usd"`
	InputTokens   int64   `json:"input_tokens"`
	OutputTokens  int64   `json:"output_tokens"`
	CacheTokens   int64   `json:"cache_tokens"`
}

// SkillTrigger is one row of the Live Activity "skills triggered" table. Dead
// marks an active skill that has never fired (trigger_count_total == 0) — an
// archive candidate (spec §7.6 v_skill_triggers). The UI renders these rows
// with a warning chip when Dead is true.
type SkillTrigger struct {
	Name        string `json:"name"`
	Count       int    `json:"count"`
	LastFiredMs int64  `json:"last_fired_ms"`
	Dead        bool   `json:"dead"`
}

// SkillStatRow is one raw row behind the Skills Analytics screen: a resolved
// skill (active or disabled) joined to its cumulative trigger count. It is the
// pure input to skills.BuildAnalytics, which derives the value ratio and the
// dead / over-triggering / disabled flags. EstContextTokens is the skill's
// always-on context tax (0 for disabled skills — non-active items estimate 0).
type SkillStatRow struct {
	Name             string
	EffectiveStatus  EffectiveStatus
	EstContextTokens int
	Triggers         int
	LastFiredMs      int64
}

// SkillAnalyticsItem is one row of the Skills Analytics screen: a skill plus its
// derived value ratio and hygiene flags. ValueRatio is triggers per 1,000 tokens
// of always-on context cost (0 when the skill costs 0 tokens). Dead = active but
// never fired; OverTriggering = fires a lot yet still poor value; Disabled =
// turned off in settings.
type SkillAnalyticsItem struct {
	Name             string  `json:"name"`
	Triggers         int     `json:"triggers"`
	LastFiredMs      int64   `json:"last_fired_ms"`
	EstContextTokens int     `json:"est_context_tokens"`
	ValueRatio       float64 `json:"value_ratio"`
	Dead             bool    `json:"dead"`
	OverTriggering   bool    `json:"over_triggering"`
	Disabled         bool    `json:"disabled"`
}

// SkillCounts tallies how many skills carry each hygiene flag (for the screen's
// summary header).
type SkillCounts struct {
	Dead           int `json:"dead"`
	OverTriggering int `json:"over_triggering"`
	Disabled       int `json:"disabled"`
}

// SkillsSnapshot is the read-model served by GET /api/skills. Items is always
// non-nil (empty slice, not null) for a stable JSON shape. TotalContextTokens
// is the summed always-on tax of all listed skills — the headline number.
type SkillsSnapshot struct {
	Items              []SkillAnalyticsItem `json:"items"`
	Counts             SkillCounts          `json:"counts"`
	Total              int                  `json:"total"`
	TotalContextTokens int                  `json:"total_context_tokens"`
}

// CounterSet holds the 5 live-activity tallies for one time window / session.
// It is returned by Store.ActivityCounters and consumed by the activity service
// layer (Tasks 9–10) which marshals it to JSON for the UI.
type CounterSet struct {
	Prompts int `json:"prompts"`
	Skills  int `json:"skills"`
	Tools   int `json:"tools"`
	Blocked int `json:"blocked"`
	Errors  int `json:"errors"`
}

// RecentEvent is one row in the live ticker. The three core fields are always
// populated; ToolName/SkillName/Status are set only for event types that carry
// them (tool_use, skill, blocked/error respectively) and are omitted from JSON
// when empty so prompt rows stay compact (Privacy D8: no raw text, names only).
// ID is the local cache row id (a meaningless monotonic counter, not sensitive)
// — the UI needs it as a stable, unique key for the live-stream list, since two
// events can share the same ts_ms+type+session_id.
type RecentEvent struct {
	ID        int64  `json:"id"`
	TsMs      int64  `json:"ts_ms"`
	Type      string `json:"type"`
	SessionID string `json:"session_id"`
	ToolName  string `json:"tool_name,omitempty"`
	SkillName string `json:"skill_name,omitempty"`
	Status    string `json:"status,omitempty"`
}

// ActivityCounters pairs the current-session tallies with the rolling 24h tallies.
// The UI shows both per counter card so users can compare their current session
// to their recent daily baseline at a glance.
type ActivityCounters struct {
	Session CounterSet `json:"session"`
	Last24h CounterSet `json:"last_24h"`
}

// Sparklines holds the per-minute event-rate series shown in the live stream
// header. Each slice has exactly sparkBuckets elements (30 by default), ordered
// oldest→newest. Zero elements represent minutes with no activity.
type Sparklines struct {
	PromptsPerMin []int `json:"prompts_per_min"`
	SkillsPerMin  []int `json:"skills_per_min"`
}

// ActivitySnapshot is the single Live Activity payload shared by the REST
// endpoint (Task 10) and the SSE broadcast (Task 12), ensuring the initial page
// load and the live stream are always in sync — they both come from the same
// assembly function.
type ActivitySnapshot struct {
	Counters   ActivityCounters `json:"counters"`
	Recent     []RecentEvent    `json:"recent"`
	Skills     []SkillTrigger   `json:"skills"`
	Sparklines Sparklines       `json:"sparklines"`
}

// Scope is where a config item was discovered. These are NOT a global
// precedence order — each category defines its own (see internal/resolve).
// Module 1 parsers emit only ScopeUser and ScopeProject; the engine handles
// every value so table tests can exercise all orders.
type Scope string

const (
	// ScopeEnterprise represents Claude's "managed" scope.
	ScopeEnterprise Scope = "enterprise"
	// ScopeLocal represents a local scope.
	ScopeLocal Scope = "local"
	// ScopeProject represents a project scope.
	ScopeProject Scope = "project"
	// ScopeUser represents a user scope.
	ScopeUser Scope = "user"
	// ScopePlugin represents a plugin scope.
	ScopePlugin Scope = "plugin"
	// ScopeBundled represents a bundled scope.
	ScopeBundled Scope = "bundled"
)

// Category is a kind of harness component. Module 1 covers the core 4.
type Category string

const (
	// CatSkill represents a skill component.
	CatSkill Category = "skill"
	// CatMCP represents an MCP component.
	CatMCP Category = "mcp"
	// CatHook represents a hook component.
	CatHook Category = "hook"
	// CatAgent represents an agent component.
	CatAgent Category = "agent"
	// CatMemory represents a CLAUDE.md / rules memory file.
	CatMemory Category = "memory"
	// CatCommand represents a custom slash command.
	CatCommand Category = "command"
	// CatOutputStyle represents an output style.
	CatOutputStyle Category = "output-style"
	// CatPlugin represents an enabled/disabled plugin.
	CatPlugin Category = "plugin"
)

// EffectiveStatus is the resolved outcome shown as a status chip.
type EffectiveStatus string

const (
	// StatusActive represents an active component.
	StatusActive EffectiveStatus = "active"
	// StatusOverridden represents an overridden component.
	StatusOverridden EffectiveStatus = "overridden"
	// StatusDisabled represents a disabled component.
	StatusDisabled EffectiveStatus = "disabled"
	// StatusShadowed represents a shadowed component.
	StatusShadowed EffectiveStatus = "shadowed"
)

// InventoryItem is one raw discovered component, before resolution. Attrs holds
// category-specific fields (skill: description/allowed_tools; agent: model;
// hook: event/matcher/command; mcp: transport/command) so the table stays flat.
type InventoryItem struct {
	AgentCode   string // which agent owns this (e.g. "claude")
	ProjectRoot string // "" = user-global (no project context)
	Category    Category
	Name        string
	Scope       Scope
	RelPath     string // relative to the scope root
	Enabled     bool   // false if a source toggle disabled it
	Attrs       map[string]string
}

// PrecedenceStep is one entry in the "why?" trail (spec §9). We record scope +
// decision + a human reason; item ids are not needed by the UI.
type PrecedenceStep struct {
	Step     int    `json:"step"`
	Scope    string `json:"scope"`
	Decision string `json:"decision"` // found|wins|overridden|disabled|shadowed
	Reason   string `json:"reason"`
}

// ResolvedItem is the materialized precedence outcome for one (category, name).
// Winner points at the raw item that took effect (nil when fully disabled).
type ResolvedItem struct {
	AgentCode        string
	ProjectRoot      string
	Category         Category
	Name             string
	EffectiveStatus  EffectiveStatus
	Winner           *InventoryItem // raw item that took effect; nil when fully disabled
	PrecedenceTrail  []PrecedenceStep
	EstContextTokens int // rough character count / 4 estimate
}

// ResolvedRow is the read-model the API serves for the Inventory table + drawer.
// It is shaped by v_active_inventory, not the physical tables.
type ResolvedRow struct {
	ID               int64             `json:"id"`
	Category         string            `json:"category"`
	Name             string            `json:"name"`
	EffectiveStatus  string            `json:"effective_status"`
	WinnerScope      string            `json:"winner_scope"`
	WinnerPath       string            `json:"winner_path"`
	InUser           bool              `json:"in_user"`
	InProject        bool              `json:"in_project"`
	EstContextTokens int               `json:"est_context_tokens"`
	Attrs            map[string]string `json:"attrs"`
}

// Toggles are the explicit on/off inputs parsed from settings*.json that the
// resolve engine consumes alongside discovered items.
type Toggles struct {
	DisableBundledSkills   bool
	SkillOverrides         map[string]string // name -> on|name-only|off
	DisabledMcpjsonServers []string
	EnabledMcpjsonServers  []string // nil = no allowlist; non-nil = allowlist active
	EnableAllProjectMcp    bool
	EnabledPlugins         []string
	// OutputStyle is the selected output style (the `outputStyle` setting). ""
	// means none set, which Claude treats as the built-in "Default".
	OutputStyle string
	// ClaudeMdExcludes are glob patterns (absolute-path) that suppress matching
	// CLAUDE.md / rules files from being loaded into context.
	ClaudeMdExcludes []string
}

// DailyUsage is one day's token + cost totals for the trend chart. Day is local
// yyyymmdd; Label is the short weekday ("Mon"); CostUSD is the summed estimate.
type DailyUsage struct {
	Day          int     `json:"day"`
	Label        string  `json:"label"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	CacheTokens  int64   `json:"cache_tokens"`
	TotalTokens  int64   `json:"total_tokens"`
	CostUSD      float64 `json:"cost_usd"`
}

// ProjectCost is the store-level grouping of cost by raw project key (the encoded
// transcript dir). The services layer relabels Root into a human Name.
type ProjectCost struct {
	Root    string
	CostUSD float64
}

// ProjectUsage is one row of the "By project" breakdown. Name is the display
// label (last '-' segment of the encoded root); Pct is the bar width relative to
// the largest project's cost in the window.
type ProjectUsage struct {
	Name    string  `json:"name"`
	CostUSD float64 `json:"cost_usd"`
	Pct     int     `json:"pct"`
}

// TokensByModel is the store-level grouping of total tokens by model. (Named to
// avoid the model.ModelCost stutter that revive's exported rule rejects.)
type TokensByModel struct {
	Model       string
	TotalTokens int64
}

// UsageShare is one row of the "By model" breakdown — a labelled percentage. Pct
// is the share of total tokens in the window. (Named to avoid the
// model.ModelUsage stutter that revive's exported rule rejects.)
type UsageShare struct {
	Name string `json:"name"`
	Pct  int    `json:"pct"`
}

// HeatDay is one cell of the activity heatmap. Bucket is an intensity level 0-3
// (0 = no activity) computed relative to the busiest day in the heatmap window.
type HeatDay struct {
	Day         int   `json:"day"`
	TotalTokens int64 `json:"total_tokens"`
	Bucket      int   `json:"bucket"`
}

// UsageSnapshot is the single payload the /api/usage endpoint serves. WindowDays
// is the trend/breakdown window (7); Heatmap spans a longer fixed window (see the
// usage service). Estimate is always true for subscription plans (API-equivalent).
type UsageSnapshot struct {
	WindowDays   int            `json:"window_days"`
	Days         []DailyUsage   `json:"days"`
	TotalCostUSD float64        `json:"total_cost_usd"`
	TotalTokens  int64          `json:"total_tokens"`
	ByProject    []ProjectUsage `json:"by_project"`
	ByModel      []UsageShare   `json:"by_model"`
	Heatmap      []HeatDay      `json:"heatmap"`
	StreakDays   int            `json:"streak_days"`
	Estimate     bool           `json:"estimate"`
}

// QuotaWindow is the latest plan-quota reading for one window (five_hour or
// seven_day). UsedPercentage is 0-100; ResetsAtMs is when the window resets.
type QuotaWindow struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAtMs     int64   `json:"resets_at_ms"`
	TsMs           int64   `json:"ts_ms"`
}

// QuotaSnapshot is the /api/quota payload + SSE "quota" frame. Available is false
// when no sample has ever been received (the UI renders the gated state). The two
// window pointers are nil when that window has no sample.
type QuotaSnapshot struct {
	Available bool         `json:"available"`
	Plan      string       `json:"plan,omitempty"`
	Source    string       `json:"source,omitempty"`
	FiveHour  *QuotaWindow `json:"five_hour"`
	SevenDay  *QuotaWindow `json:"seven_day"`
}

// QuotaSampleRow is one quota reading to persist (one row per window). It is the
// input to Store.InsertQuotaSample.
type QuotaSampleRow struct {
	AgentCode      string
	Window         string
	UsedPercentage float64
	ResetsAtMs     int64
	TsMs           int64
	Plan           string
	Source         string
}

// QuotaWindowRow is one row read back from v_latest_quota by Store.LatestQuota.
type QuotaWindowRow struct {
	Window         string
	UsedPercentage float64
	ResetsAtMs     int64
	TsMs           int64
	Plan           string
	Source         string
}

// ContextBudgetSnapshot is the Context-Budget screen payload: the always-on
// "context tax" total, the configured window denominator, the per-category
// breakdown (stacked bar), every active consumer (biggest-consumers table +
// the client-side "if disabled" recompute source), and honesty caveats.
type ContextBudgetSnapshot struct {
	TotalTokens  int              `json:"total_tokens"`
	WindowTokens int              `json:"window_tokens"`
	Pct          float64          `json:"pct"`
	ByCategory   []CategoryBudget `json:"by_category"`
	Consumers    []ConsumerItem   `json:"consumers"`
	Caveats      []string         `json:"caveats"`
}

// CategoryBudget is one stacked-bar segment: a category's summed tokens, the
// number of active items in it, and its share of the total tax.
type CategoryBudget struct {
	Category string  `json:"category"`
	Tokens   int     `json:"tokens"`
	Count    int     `json:"count"`
	Pct      float64 `json:"pct"`
}

// ConsumerItem is one active component in the biggest-consumers table; the
// frontend subtracts Tokens when the user toggles it off in the recompute.
type ConsumerItem struct {
	ID       int64  `json:"id"`
	Category string `json:"category"`
	Name     string `json:"name"`
	Scope    string `json:"scope"`
	Tokens   int    `json:"tokens"`
}

// Severity ranks a security finding (high > medium > low). It drives the
// screen's colour coding (red / amber / grey) and sort order.
type Severity string

const (
	// SeverityHigh marks an issue to act on now.
	SeverityHigh Severity = "high"
	// SeverityMedium marks an issue that should be fixed.
	SeverityMedium Severity = "medium"
	// SeverityLow marks an issue to be aware of.
	SeverityLow Severity = "low"
)

// ValidSeverity reports whether s is one of the three known severities. The
// rules loader uses it to drop rules with an unrecognised severity.
func ValidSeverity(s string) bool {
	switch Severity(s) {
	case SeverityHigh, SeverityMedium, SeverityLow:
		return true
	}
	return false
}

// ScopePermissions is the permission input parsed from ONE settings file. The
// engine reads these to detect missing denies, broad allows, and risky modes.
type ScopePermissions struct {
	Scope             Scope
	RelPath           string
	Deny              []string
	Allow             []string
	Ask               []string
	DefaultMode       string
	SecretSettingKeys []string // settings keys whose value looked like a secret; values are scrubbed
}

// MCPEnvShape is the privacy-safe shape of one MCP server's env block: the names
// of keys whose values looked like secrets. Values themselves are never captured.
type MCPEnvShape struct {
	Server     string
	Scope      Scope
	RelPath    string
	SecretKeys []string
}

// PluginSource identifies one enabled plugin and the marketplace it came from.
type PluginSource struct {
	Name        string
	Marketplace string
	Scope       Scope
	RelPath     string
}

// SecurityInputs is everything the security rule engine needs from one discovery
// location. It is kept separate from Toggles: Toggles drives enable/disable
// resolution, SecurityInputs drives the audit.
type SecurityInputs struct {
	Permissions []ScopePermissions
	MCPEnv      []MCPEnvShape
	Plugins     []PluginSource
}

// Finding is one security issue emitted by the rule engine. Detail names the
// offending key only — it NEVER contains a secret value.
type Finding struct {
	RuleID      string `json:"rule_id"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	TargetKey   string `json:"target_key"`
	Detail      string `json:"detail"`
	Remediation string `json:"remediation"`
	Scope       string `json:"scope"`
}

// SecuritySnapshot is the read-model the API serves for the Security screen:
// the findings plus per-severity counts and a total.
type SecuritySnapshot struct {
	Findings []Finding      `json:"findings"`
	Counts   map[string]int `json:"counts"`
	Total    int            `json:"total"`
}

// Alert is one Overview alert row, derived fresh each cycle (no persistence): it
// is present while its condition holds and disappears when it clears. TsMs is set
// only for event-derived alerts (e.g. a blocked command) and omitted for
// current-state alerts. CTA is a UI route key. M8 spec §7.
type Alert struct {
	Kind     string `json:"kind"`
	Severity string `json:"severity"`
	Text     string `json:"text"`
	TsMs     int64  `json:"ts_ms,omitempty"`
	CTA      string `json:"cta"`
}

// ActiveComponents is the per-category active-item census on the Overview KPI
// card. Counts are active-only, scoped to one root. M8 spec §4.
type ActiveComponents struct {
	Total      int             `json:"total"`
	ByCategory []CategoryCount `json:"by_category"`
}

// CategoryCount is one category's active count plus its user/project scope split.
type CategoryCount struct {
	Category     string `json:"category"`
	Count        int    `json:"count"`
	UserCount    int    `json:"user_count"`
	ProjectCount int    `json:"project_count"`
}

// ActiveHereRow is one deep-linkable line in the Overview "Active here" panel: a
// category count, an honest scope note, and the route to open. M8 spec §4.
type ActiveHereRow struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
	Note     string `json:"note"`
	CTA      string `json:"cta"`
}

// ContextTax is the compact Overview surfacing of the always-on context cost
// (the number M4 deferred to M8). M8 spec §5.
type ContextTax struct {
	TotalTokens  int     `json:"total_tokens"`
	WindowTokens int     `json:"window_tokens"`
	Pct          float64 `json:"pct"`
}

// HealthBar is one labelled sub-score (0–100) in the harness-health composite.
type HealthBar struct {
	Label string `json:"label"`
	Score int    `json:"score"`
}

// HealthSnapshot is the Overview harness-health composite: a 0–100 ring Score
// plus the four sub-score Bars it averages (context-tax, security, skill-hygiene,
// hook-health). See M8 spec §6.
type HealthSnapshot struct {
	Score int         `json:"score"`
	Bars  []HealthBar `json:"bars"`
}
