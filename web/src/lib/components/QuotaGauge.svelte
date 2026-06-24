<!--
  QuotaGauge.svelte — one plan-quota ring (mockup: SVG circle with dashoffset).

  Props:
    window   — { used_percentage, resets_at_ms } | null. null → gated state.
    label    — e.g. "Session · 5h window"
    subtitle — e.g. "rate_limits.five_hour"
    color    — ring stroke colour (CSS var); default amber.

  Gated state: when window is null we render an install call-to-action instead of
  a number, so the screen is honest before the statusline helper is installed.
-->
<script lang="ts">
  import type { QuotaWindow } from '$lib/api';

  let {
    window = null,
    label = '',
    subtitle = '',
    color = 'var(--amber)',
  }: {
    window?: QuotaWindow | null;
    label?: string;
    subtitle?: string;
    color?: string;
  } = $props();

  // The mockup ring circumference for r=42 is 2*pi*42 ≈ 263.9.
  const CIRC = 263.9;
  // pctVal: the used percentage clamped to 0-100.
  const pctVal = $derived(Math.max(0, Math.min(100, window?.used_percentage ?? 0)));
  // dash: stroke-dashoffset that fills the ring proportional to usage.
  const dash = $derived((CIRC * (1 - pctVal / 100)).toFixed(1));
  // resets: human "resets in" text from resets_at_ms (re-uses the ago formatter
  // inverted — we show time remaining as a coarse string).
  const resetsTxt = $derived(window ? msUntil(window.resets_at_ms) : '');

  // msUntil renders a coarse "in 1h 12m" / "in 3d" until a future epoch-ms.
  function msUntil(ms: number): string {
    const s = Math.max(0, Math.floor((ms - Date.now()) / 1000));
    if (s >= 86400) return `${Math.floor(s / 86400)}d`;
    if (s >= 3600) return `${Math.floor(s / 3600)}h ${Math.floor((s % 3600) / 60)}m`;
    if (s >= 60) return `${Math.floor(s / 60)}m`;
    return `${s}s`;
  }
</script>

<div class="gauge">
  {#if window}
    <div class="ring">
      <svg viewBox="0 0 100 100">
        <circle cx="50" cy="50" r="42" fill="none" stroke="var(--border)" stroke-width="9" />
        <circle
          cx="50" cy="50" r="42" fill="none" stroke={color} stroke-width="9"
          stroke-linecap="round" stroke-dasharray={CIRC} stroke-dashoffset={dash}
        />
      </svg>
      <span class="num">{Math.round(pctVal)}%</span>
    </div>
    <div class="meta">
      <div class="label">{label}</div>
      <div class="resets">Resets in <span>{resetsTxt}</span></div>
      <div class="sub">From statusline <code>{subtitle}</code></div>
    </div>
  {:else}
    <!-- Gated: no sample yet. -->
    <div class="ring gated">
      <svg viewBox="0 0 100 100">
        <circle cx="50" cy="50" r="42" fill="none" stroke="var(--border)" stroke-width="9" />
      </svg>
      <span class="num muted">–</span>
    </div>
    <div class="meta">
      <div class="label">{label}</div>
      <div class="resets muted">Install the statusline helper</div>
      <div class="sub">to see live <code>{subtitle}</code></div>
    </div>
  {/if}
</div>

<style>
  .gauge { display: flex; align-items: center; gap: 18px; border: 1px solid var(--border); border-radius: 11px; background: var(--panel); padding: 16px 18px; }
  .ring { position: relative; width: 96px; height: 96px; flex: none; }
  .ring svg { width: 96px; height: 96px; transform: rotate(-90deg); }
  .ring circle[stroke-linecap='round'] { transition: stroke-dashoffset 0.6s ease; }
  .num { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; font-size: 24px; font-weight: 600; font-variant-numeric: tabular-nums; }
  .num.muted { color: var(--text-faint); }
  .label { font-size: 12px; color: var(--text-faint); text-transform: uppercase; letter-spacing: 0.05em; }
  .resets { font-size: 15px; font-weight: 600; margin: 5px 0 3px; }
  .resets.muted { color: var(--text-dim); font-weight: 500; }
  .sub { font-size: 12px; color: var(--text-dim); }
  code { font-family: 'IBM Plex Mono', monospace; }
</style>
