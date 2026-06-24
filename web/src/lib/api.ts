// Typed fetch wrapper for the JSON API. Throws on the uniform error envelope
// so callers render the inline error/retry UI state (spec §12.2).
// All field names mirror the Go model JSON tags exactly (snake_case).

/** ResolvedRow is the read-model the API returns for each inventory item. */
export type ResolvedRow = {
  id: number;
  category: string;
  name: string;
  effective_status: string;   // "active" | "overridden" | "disabled" | "shadowed"
  winner_scope: string;
  winner_path: string;
  in_user: boolean;
  in_project: boolean;
  est_context_tokens: number;
  attrs: Record<string, string>;
};

/** TrailStep is one entry in the "why?" precedence trail for a resolved row. */
export type TrailStep = {
  step: number;
  scope: string;
  decision: string; // found | wins | overridden | disabled | shadowed
  reason: string;
};

// get: internal helper — fetches a JSON endpoint and throws on the uniform
// error envelope { error: { message } } returned by the Go API on failures.
async function get<T>(url: string): Promise<T> {
  const r = await fetch(url);
  const body = await r.json();
  // The Go API always uses this envelope on error (spec §11).
  if (body?.error) throw new Error(body.error.message);
  return body as T;
}

/**
 * getInventory fetches the resolved inventory rows for a given category.
 * showDisabled=true includes disabled and shadowed rows (show_disabled=1).
 */
export const getInventory = (category: string, showDisabled: boolean) =>
  get<{ items: ResolvedRow[] }>(
    `/api/inventory?category=${encodeURIComponent(category)}&show_disabled=${showDisabled ? 1 : 0}`
  );

/**
 * getWhy fetches the precedence trail for a single resolved row by its id.
 * Used by the "why?" drawer to explain how a row reached its effective status.
 */
export const getWhy = (id: number) =>
  get<{ trail: TrailStep[] }>(`/api/inventory/${id}/why`);

/** DailyUsage mirrors Go model.DailyUsage. */
export type DailyUsage = {
  day: number;
  label: string;
  input_tokens: number;
  output_tokens: number;
  cache_tokens: number;
  total_tokens: number;
  cost_usd: number;
};

/** UsageSnapshot mirrors Go model.UsageSnapshot. */
export type UsageSnapshot = {
  window_days: number;
  days: DailyUsage[];
  total_cost_usd: number;
  total_tokens: number;
  by_project: { name: string; cost_usd: number; pct: number }[];
  by_model: { name: string; pct: number }[];
  heatmap: { day: number; total_tokens: number; bucket: number }[];
  streak_days: number;
  estimate: boolean;
};

/** QuotaWindow mirrors Go model.QuotaWindow. */
export type QuotaWindow = { used_percentage: number; resets_at_ms: number; ts_ms: number };

/** QuotaSnapshot mirrors Go model.QuotaSnapshot (null windows when no sample). */
export type QuotaSnapshot = {
  available: boolean;
  plan?: string;
  source?: string;
  five_hour: QuotaWindow | null;
  seven_day: QuotaWindow | null;
};

/** getUsage fetches the Usage & Cost snapshot (fixed 7-day window). */
export const getUsage = () => get<UsageSnapshot>('/api/usage');

/** getQuota fetches the live plan-quota snapshot (gated when no helper installed). */
export const getQuota = () => get<QuotaSnapshot>('/api/quota');

/** CategoryBudget mirrors Go model.CategoryBudget (one stacked-bar segment). */
export type CategoryBudget = {
  category: string;
  tokens: number;
  count: number;
  pct: number;
};

/** ConsumerItem mirrors Go model.ConsumerItem (one biggest-consumers row). */
export type ConsumerItem = {
  id: number;
  category: string;
  name: string;
  scope: string;
  tokens: number;
};

/** ContextBudgetSnapshot mirrors Go model.ContextBudgetSnapshot. */
export type ContextBudgetSnapshot = {
  total_tokens: number;
  window_tokens: number;
  pct: number;
  by_category: CategoryBudget[];
  consumers: ConsumerItem[];
  caveats: string[];
};

/** getContextBudget fetches the always-on context-tax snapshot. */
export const getContextBudget = () => get<ContextBudgetSnapshot>('/api/context-budget');

/** Finding mirrors Go model.Finding. */
export type Finding = {
  rule_id: string;
  severity: 'high' | 'medium' | 'low';
  title: string;
  target_key: string;
  detail: string;
  remediation: string;
  scope: string;
};

/** SecuritySnapshot mirrors Go model.SecuritySnapshot. */
export type SecuritySnapshot = {
  findings: Finding[];
  counts: Record<string, number>;
  total: number;
};

/** getSecurity fetches the Security & Audit findings snapshot. */
export const getSecurity = () => get<SecuritySnapshot>('/api/security');

/** SkillItem mirrors Go model.SkillAnalyticsItem (one Skills-screen row). */
export type SkillItem = {
  name: string;
  triggers: number;
  last_fired_ms: number;
  est_context_tokens: number;
  value_ratio: number;
  dead: boolean;
  over_triggering: boolean;
  disabled: boolean;
};

/** SkillsSnapshot mirrors Go model.SkillsSnapshot. */
export type SkillsSnapshot = {
  items: SkillItem[];
  counts: { dead: number; over_triggering: number; disabled: number };
  total: number;
  total_context_tokens: number;
};

/** getSkills fetches the Skills Analytics snapshot. */
export const getSkills = () => get<SkillsSnapshot>('/api/skills');

// ── Settings types (Module 7) ───────────────────────────────────────────────

/**
 * SettingsView mirrors Go settings.View: the full GET /api/settings response.
 * Duration fields (active_window, aggregate_horizon, throttle, check_interval)
 * come back as Go duration strings (e.g. "48h0m0s") — the UI displays and
 * sends them as-is so the daemon parses them.
 */
export type SettingsView = {
  // Editable
  port: number;
  bind_addr: string;
  theme: string;
  accent: string;
  active_window: string;
  aggregate_horizon: string;
  throttle: string;
  check_interval: string;
  context_window_tokens: number;
  roots: string[];
  auto_check: boolean;
  // Read-only
  version: string;
  scrub_locked: boolean;
  outbound_default_off: boolean;
  db_bytes: number;
  backup_bytes: number;
  mcp_servers: string[];
};

/**
 * SettingsSaveResult is the PUT /api/settings response.
 * restart_required is true when port or bind_addr changed (daemon restart needed).
 * warnings is a list of non-fatal advisory strings (e.g. clamping notifications).
 */
export type SettingsSaveResult = {
  saved: boolean;
  restart_required: boolean;
  warnings: string[];
};

/**
 * SettingsInput mirrors Go settings.Input: the PUT /api/settings body.
 * All fields are optional — the server treats zero/empty as "no change".
 * auto_check is explicitly typed as boolean | undefined so the caller can
 * omit it (no-change) or pass false (explicit disable).
 */
export type SettingsInput = {
  port?: number;
  bind_addr?: string;
  theme?: string;
  accent?: string;
  active_window?: string;
  aggregate_horizon?: string;
  throttle?: string;
  check_interval?: string;
  context_window_tokens?: number;
  auto_check?: boolean;
};

/**
 * DirsResult is the GET /api/roots?path= response.
 * home is the user's home directory; dirs are immediate subdirectories.
 */
export type DirsResult = { home: string; dirs: string[] };

/**
 * UpdateStatus is the GET /api/update/status?check=1 response.
 * available is true when latest > current. commands is the upgrade recipe.
 */
export type UpdateStatus = {
  current: string;
  latest: string;
  available: boolean;
  commands: string[];
};

/**
 * StatuslineResult is the GET /api/install/statusline response.
 * proposed is the full suggested settings.json content; added lists changed keys;
 * path is the absolute path to the user's settings.json.
 */
export type StatuslineResult = {
  proposed: string;
  added: string[];
  path: string;
};

/**
 * SecurityRule mirrors Go security.Rule: one rule in the rules array.
 * Array fields (patterns, modes, allowed, keywords, prefixes) are optional and
 * omitted when empty — matching the Go `json:"...,omitempty"` tags.
 */
export type SecurityRule = {
  id: string;
  type: string;
  enabled: boolean;
  severity: string;
  title: string;
  remediation: string;
  patterns?: string[];
  modes?: string[];
  allowed?: string[];
  keywords?: string[];
  prefixes?: string[];
};

/**
 * Thresholds mirrors Go skills.Thresholds: the two skill-analytics knobs.
 * high_trigger_min: fire count before a low-value skill is flagged over-triggering.
 * low_value_ratio_max: value ratio below which a heavy skill is considered low-value.
 */
export type Thresholds = {
  high_trigger_min: number;
  low_value_ratio_max: number;
};

// ── Settings API helpers ────────────────────────────────────────────────────

// send: internal helper for PUT/POST with a JSON body. Mirrors get<T> in that
// it throws on the uniform error envelope { error: { message } } returned by
// the Go API on failures. Returns the parsed response body as T.
async function send<T>(url: string, method: string, body: unknown): Promise<T> {
  const r = await fetch(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  const b = await r.json();
  // The Go API always uses { error: { message } } on 4xx/5xx.
  if (b?.error) throw new Error(b.error.message);
  return b as T;
}

/** getSettings fetches the current settings view (GET /api/settings). */
export const getSettings = () => get<SettingsView>('/api/settings');

/**
 * putSettings sends an input object to PUT /api/settings and returns the save
 * result (saved, restart_required, warnings). Send the FULL current state on
 * every save so nothing is clobbered — the server's "empty = no change"
 * semantics only apply to fields that are genuinely absent or zero.
 */
export const putSettings = (body: SettingsInput) =>
  send<SettingsSaveResult>('/api/settings', 'PUT', body);

/** listDirs browses immediate subdirectories under path (GET /api/roots?path=). */
export const listDirs = (path: string) =>
  get<DirsResult>(`/api/roots?path=${encodeURIComponent(path)}`);

/** setRoots saves the roots list (PUT /api/roots). */
export const setRoots = (paths: string[]) =>
  send<{ saved: boolean }>('/api/roots', 'PUT', { paths });

/** ActiveRoot is the top-bar selector's view: the current scope ("" = All), the
 *  daemon's primary folder (info), and every configured folder. The "All" option
 *  is the empty string and is added client-side. */
export type ActiveRoot = { current: string; default: string; roots: string[] };

/** getActiveRoot fetches the selectable roots + current selection (GET /api/active-root). */
export const getActiveRoot = () => get<ActiveRoot>('/api/active-root');

/** setActiveRoot switches the global view root (PUT /api/active-root). */
export const setActiveRoot = (root: string) =>
  send<{ current: string }>('/api/active-root', 'PUT', { root });

/**
 * checkUpdate checks for a new Drishti version (GET /api/update/status?check=1).
 * check=1 forces a live fetch rather than returning a cached result.
 */
export const checkUpdate = () => get<UpdateStatus>('/api/update/status?check=1');

/**
 * proposeStatusline fetches a non-mutating statusLine suggestion
 * (GET /api/install/statusline). The daemon NEVER writes the user's file.
 */
export const proposeStatusline = () => get<StatuslineResult>('/api/install/statusline');

/** getThresholds fetches the daemon's current skill-analytics thresholds (GET /api/thresholds). */
export const getThresholds = () => get<Thresholds>('/api/thresholds');

/** putThresholds saves skills-analytics.toml thresholds (PUT /api/thresholds). */
export const putThresholds = (body: Thresholds) =>
  send<{ saved: boolean }>('/api/thresholds', 'PUT', body);

/** putRules saves security-rules.toml from a rules array (PUT /api/rules). */
export const putRules = (rules: SecurityRule[]) =>
  send<{ saved: boolean }>('/api/rules', 'PUT', rules);
