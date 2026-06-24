<!--
  Usage & Cost — token/cost analytics from transcripts + live plan-quota gauges.

  Data model:
    • Usage snapshot is fetched once on mount (it changes only every few minutes;
      no dedicated SSE frame, per the M3 spec).
    • Quota is live: the shared EventSource feeds the `quota` store; we also fetch
      /api/quota on mount so the gauges populate before the first SSE frame.
  States: loading (no usage yet), gated (quota unavailable), loaded.
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import { getUsage, getQuota, type UsageSnapshot } from '$lib/api';
  import { quota, rootVersion } from '$lib/sse';
  import { usd } from '$lib/format';
  import QuotaGauge from '$lib/components/QuotaGauge.svelte';
  import TrendChart from '$lib/components/TrendChart.svelte';
  import Heatmap from '$lib/components/Heatmap.svelte';

  // usage: the snapshot, null until the mount fetch resolves.
  let usage = $state<UsageSnapshot | null>(null);
  // err: inline error string if the fetch fails (the page shows it with a retry).
  let err = $state<string | null>(null);

  async function load() {
    err = null;
    try {
      usage = await getUsage();
    } catch (e) {
      err = e instanceof Error ? e.message : 'failed to load usage';
    }
  }

  onMount(() => {
    load();
    // Seed the quota store before the first SSE frame (best-effort).
    getQuota().then((q) => quota.set(q)).catch(() => {});
  });

  // Re-fetch when the top-bar root selector switches the active root (bump > 0).
  $effect(() => {
    const _v = $rootVersion;
    if (_v > 0) load();
  });

  // q: the live quota snapshot (store), or null.
  const q = $derived($quota);
</script>

<div class="head">
  <h1>Usage &amp; Cost</h1>
  <p>Token &amp; cost analytics from transcripts, plus the live plan-quota gauge.</p>
</div>

<!-- API-equivalent estimate caveat (subscription plans aren't billed per-token). -->
<div class="caveat">
  <span class="i">ⓘ</span>
  Subscription plan: figures are an API-equivalent <strong>estimate</strong>, not billed spend.
</div>

<!-- Quota gauges (gated until the statusline helper is installed). -->
<div class="gauges">
  <QuotaGauge window={q?.five_hour ?? null} label="Session · 5h window" subtitle="rate_limits.five_hour" color="var(--amber)" />
  <QuotaGauge window={q?.seven_day ?? null} label="Weekly window" subtitle="rate_limits.seven_day" color="var(--accent)" />
</div>

{#if err}
  <div class="error">Couldn't load usage: {err} <button onclick={load}>Retry</button></div>
{:else if !usage}
  <p class="empty">Loading usage…</p>
{:else}
  <TrendChart days={usage.days} totalCost={usage.total_cost_usd} />

  <div class="grid">
    <!-- By project -->
    <div class="panel">
      <div class="panel-head">By project</div>
      <!-- Key by index: this is a static list fetched once (no live reordering),
           and two different project roots can share a display label, which would
           collide on (p.name) and crash the render with each_key_duplicate. -->
      {#each usage.by_project as p, i (i)}
        <div class="row">
          <span class="name mono">{p.name}</span>
          <span class="bar"><span style="width:{p.pct}%;background:var(--accent)"></span></span>
          <span class="val">{usd(p.cost_usd)}</span>
        </div>
      {/each}
      {#if usage.by_project.length === 0}<div class="row empty">No usage yet.</div>{/if}
    </div>

    <!-- By model -->
    <div class="panel">
      <div class="panel-head">By model</div>
      <!-- Key by index: static list, rendered once. The backend now merges models
           by label so names are unique, but index keys keep the render crash-proof
           even if an unmapped id ever repeats a label. -->
      {#each usage.by_model as m, i (i)}
        <div class="row">
          <span class="name">{m.name}</span>
          <span class="bar"><span style="width:{m.pct}%;background:var(--accent-dim)"></span></span>
          <span class="val">{m.pct}%</span>
        </div>
      {/each}
      {#if usage.by_model.length === 0}<div class="row empty">No usage yet.</div>{/if}
    </div>

    <!-- Heatmap -->
    <Heatmap cells={usage.heatmap} streak={usage.streak_days} />
  </div>
{/if}

<style>
  .head { margin-bottom: 16px; }
  h1 { margin: 0; font-size: 21px; font-weight: 600; letter-spacing: -0.02em; }
  p { margin: 4px 0 0; font-size: 13px; color: var(--text-faint); }
  .caveat { display: flex; align-items: center; gap: 10px; padding: 10px 14px; border: 1px solid var(--border); border-left: 2px solid var(--amber); border-radius: 8px; background: var(--amber-soft); margin-bottom: 16px; font-size: 12.5px; color: var(--text-dim); }
  .caveat .i { color: var(--amber); }
  .gauges { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; margin-bottom: 14px; }
  .grid { display: grid; grid-template-columns: 1.2fr 1fr 1.1fr; gap: 14px; }
  .panel { border: 1px solid var(--border); border-radius: 11px; background: var(--panel); overflow: hidden; }
  .panel-head { padding: 12px 16px; border-bottom: 1px solid var(--border-soft); font-size: 12.5px; font-weight: 600; }
  .row { display: flex; align-items: center; gap: 10px; padding: 10px 16px; border-bottom: 1px solid var(--border-soft); }
  .row.empty { color: var(--text-faint); font-size: 12.5px; }
  .name { flex: 1; font-size: 12.5px; color: var(--text); }
  .mono { font-family: 'IBM Plex Mono', monospace; }
  .bar { flex: 1; height: 5px; border-radius: 3px; background: var(--border); overflow: hidden; }
  .bar span { display: block; height: 100%; }
  .val { font-size: 12.5px; font-weight: 600; font-variant-numeric: tabular-nums; width: 48px; text-align: right; }
  .empty { padding: 16px; font-size: 13px; color: var(--text-faint); }
  .error { padding: 12px 16px; font-size: 13px; color: var(--red); border: 1px solid var(--border); border-radius: 8px; margin-bottom: 14px; }
  .error button { margin-left: 8px; font: inherit; cursor: pointer; }
</style>
