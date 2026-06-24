<!--
  CounterCard.svelte — one Live Activity counter.
  Props: label, session (number), day (24h number), color (CSS var).
  Primary number = current session; secondary = last 24h.
-->
<script lang="ts">
  // Props (Svelte 5 runes style — matching Sparkline.svelte's $props() idiom).
  // $props() destructures named props with optional type annotation and defaults.
  let {
    label,
    session,
    day,
    color = 'var(--text)',
  }: {
    label: string;
    session: number;
    day: number;
    color?: string;
  } = $props();
</script>

<div class="counter">
  <div class="label">{label}</div>
  <div class="val" style="color:{color}">{session}</div>
  <div class="sub">{day} in 24h</div>
</div>

<style>
  .counter { border: 1px solid var(--border); border-radius: 10px; background: var(--panel); padding: 13px 15px; }
  .label { font-size: 11px; color: var(--text-faint); text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 7px; }
  .val { font-size: 24px; font-weight: 600; letter-spacing: -0.02em; font-variant-numeric: tabular-nums; }
  .sub { font-size: 11px; color: var(--text-faint); margin-top: 4px; font-variant-numeric: tabular-nums; }
</style>
