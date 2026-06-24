<!--
  LiveEventRow.svelte — one row in the live stream.
  Props: ev (RecentEvent). Derives icon/label/tag from event type; blocked rows
  are tinted red-soft. No raw text is present in the payload (privacy D8).
-->
<script lang="ts">
  // Reuse the shared RecentEvent type from sse.ts (already extended in Task 13).
  // Do NOT re-inline the shape here — single source of truth.
  import type { RecentEvent } from '$lib/sse';
  import { clock } from '$lib/format';

  // Props (Svelte 5 runes style — matching Sparkline.svelte's $props() idiom).
  let { ev }: { ev: RecentEvent } = $props();

  // meta: lookup table mapping event type → display icon, colour, and tag pill text.
  // This is a plain const (not reactive) because the table never changes at runtime.
  const meta: Record<string, { icon: string; color: string; tag: string }> = {
    prompt:   { icon: '▸', color: 'var(--text)',     tag: '' },
    tool_use: { icon: '⚙', color: 'var(--text-dim)', tag: '' },
    skill:    { icon: '✦', color: 'var(--accent)',   tag: 'skill' },
    blocked:  { icon: '⛔', color: 'var(--red)',      tag: 'blocked' },
    error:    { icon: '⚠', color: 'var(--amber)',    tag: 'error' },
  };

  // m: reactive derived value (Svelte 5 $derived replaces Svelte 4 `$:` statements).
  // Falls back to a neutral bullet when the event type is unrecognised.
  const m = $derived(meta[ev.type] ?? { icon: '·', color: 'var(--text-dim)', tag: '' });

  // label: human-readable row label — skill name for skill events,
  // tool name for tool_use events, or the raw type string otherwise.
  const label = $derived(
    ev.type === 'skill'    ? (ev.skill_name ?? ev.type)
    : ev.type === 'tool_use' ? (ev.tool_name  ?? ev.type)
    : ev.type
  );
</script>

<div class="row" class:blocked={ev.type === 'blocked'}>
  <span class="time">{clock(ev.ts_ms)}</span>
  <span class="icon" style="color:{m.color}">{m.icon}</span>
  <span class="label">{label}</span>
  {#if m.tag}<span class="tag tag-{m.tag}">{m.tag}</span>{/if}
</div>

<style>
  .row { display: flex; align-items: center; gap: 11px; padding: 9px; border-radius: 7px; animation: hud-in 0.25s ease; }
  .row.blocked { background: var(--red-soft); }
  .time { font-family: 'IBM Plex Mono', monospace; font-size: 11.5px; color: var(--text-faint); font-variant-numeric: tabular-nums; }
  .icon { width: 17px; text-align: center; font-size: 14px; }
  .label { font-size: 13px; color: var(--text); }
  .tag { font-size: 9.5px; font-weight: 600; padding: 1px 6px; border-radius: 5px; text-transform: uppercase; letter-spacing: 0.03em; }
  .tag-skill { color: var(--accent); background: var(--accent-soft); }
  .tag-blocked { color: var(--red); background: var(--red-soft); }
  .tag-error { color: var(--amber); background: var(--amber-soft); }
</style>
