// One EventSource for the whole app, exposed as Svelte stores. Components
// subscribe to these; they never open their own connections (spec §12.2).
import { writable } from 'svelte/store';

export type KPIs = {
  prompts_today: number;
  spend_today_usd: number;
  input_tokens: number;
  output_tokens: number;
  cache_tokens: number;
};

// Overview (Command Center) payload — mirrors Go services.Overview json tags.
export type ActiveComponents = {
  total: number;
  by_category: { category: string; count: number; user_count: number; project_count: number }[];
};
export type ActiveHereRow = { category: string; count: number; note: string; cta: string };
export type ContextTax = { total_tokens: number; window_tokens: number; pct: number };
export type HealthBar = { label: string; score: number };
export type Health = { score: number; bars: HealthBar[] };
export type Alert = { kind: string; severity: string; text: string; ts_ms?: number; cta: string };
export type Overview = {
  kpis: KPIs;
  recent: RecentEvent[];
  active_components: ActiveComponents;
  active_here: ActiveHereRow[];
  context_tax: ContextTax;
  health: Health;
  alerts: Alert[];
};

// RecentEvent matches the Go model.RecentEvent json tags (omitempty fields are
// optional here so TypeScript doesn't require them when the server omits them).
export type RecentEvent = {
  ts_ms: number;
  type: string;
  session_id: string;
  tool_name?: string;   // omitempty in Go — present only for tool_use events
  skill_name?: string;  // omitempty in Go — present only for skill events
  status?: string;      // omitempty in Go — present only when status is set
};

// CounterSet mirrors Go model.CounterSet json tags exactly.
export type CounterSet = {
  prompts: number;
  skills: number;
  tools: number;
  blocked: number;
  errors: number;
};

// SkillStat mirrors Go model.SkillStat json tags exactly.
export type SkillStat = {
  name: string;
  count: number;
  last_fired_ms: number;
  dead: boolean;
};

// ActivitySnapshot mirrors Go model.ActivitySnapshot json tags exactly.
// counters has session and last_24h buckets; sparklines are per-minute rate
// arrays for prompts and skills; recent is the rolling event log.
export type ActivitySnapshot = {
  counters: { session: CounterSet; last_24h: CounterSet };
  skills: SkillStat[];
  sparklines: { prompts_per_min: number[]; skills_per_min: number[] };
  recent: RecentEvent[];
};

export const status = writable<'live' | 'starting' | 'offline'>('starting');
export const kpis = writable<KPIs | null>(null);
export const recent = writable<RecentEvent[]>([]);

// activity: updated whenever the server broadcasts a {type:"activity"} SSE
// message. Null until the first activity snapshot arrives.
export const activity = writable<ActivitySnapshot | null>(null);

// QuotaSnapshot mirrors Go model.QuotaSnapshot json tags exactly.
export type QuotaSnapshot = {
  available: boolean;
  plan?: string;
  source?: string;
  five_hour: { used_percentage: number; resets_at_ms: number; ts_ms: number } | null;
  seven_day: { used_percentage: number; resets_at_ms: number; ts_ms: number } | null;
};

// quota: updated whenever the server broadcasts a {type:"quota"} SSE message,
// including the reconnect snapshot. Null until the first frame arrives.
export const quota = writable<QuotaSnapshot | null>(null);

// overview: the full Command Center payload, set on every "counters" frame.
export const overview = writable<Overview | null>(null);

// inventoryVersion: bumps by 1 each time an "inventory_changed" SSE event
// arrives. The Inventory page subscribes and re-fetches when the value changes.
// Using a monotonic counter (not a boolean toggle) means a rapid burst of
// changes still triggers a fetch even if Svelte batches two updates together.
export const inventoryVersion = writable<number>(0);

export function connect() {
  const es = new EventSource('/api/stream');
  es.onopen = () => status.set('live');
  es.onerror = () => status.set('offline');
  es.onmessage = (e) => {
    const m = JSON.parse(e.data);
    if (m.type === 'status') status.set(m.payload.state);
    if (m.type === 'counters') {
      overview.set(m.payload);
      kpis.set(m.payload.kpis);
      recent.set(m.payload.recent ?? []);
    }
    // activity: full ActivitySnapshot broadcast by the backend over SSE.
    if (m.type === 'activity') activity.set(m.payload);
    // quota: live plan-quota snapshot broadcast on each new sample + on reconnect.
    if (m.type === 'quota') quota.set(m.payload);
    // inventory_changed: bump the version so any subscribed page re-fetches.
    if (m.type === 'inventory_changed') {
      inventoryVersion.update((v) => v + 1);
    }
  };
  return () => es.close();
}
