<!--
  Context Budget — always-on context tax analyzer.

  Shows how many tokens the harness consumes before the user even types a
  prompt, broken down by category (stacked bar) and by individual consumers
  (table with per-row disable toggles).  A client-side "projected total"
  updates instantly as rows are toggled.

  Data model:
    • ContextBudgetSnapshot is fetched once on mount (same pattern as Usage).
    • No dedicated SSE frame — the snapshot is recomputed on demand by the API.

  States: loading (snapshot not yet arrived), error (fetch failed + retry),
          empty (total_tokens === 0), loaded.
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import { getContextBudget, type ContextBudgetSnapshot } from '$lib/api';
  import { tokensCompact, pct } from '$lib/format';

  // snapshot: the API result, null until the mount fetch resolves.
  let snapshot = $state<ContextBudgetSnapshot | null>(null);
  // err: inline error string shown with a retry button on failure.
  let err = $state<string | null>(null);

  // disabledIds: the set of consumer ids the user has "disabled" via checkboxes.
  // Default is empty — all rows start checked (enabled).
  // We track DISABLED (not enabled) so a fresh Set means "all enabled".
  let disabledIds = $state(new Set<number>());

  // load: (re-)fetches the snapshot from the API, clearing the previous error.
  async function load() {
    err = null;
    try {
      snapshot = await getContextBudget();
    } catch (e) {
      err = e instanceof Error ? e.message : 'failed to load context budget';
    }
  }

  onMount(() => { load(); });

  // projectedTokens: total_tokens minus the sum of tokens for disabled rows.
  // Updates reactively whenever disabledIds or snapshot changes.
  const projectedTokens = $derived((() => {
    if (!snapshot) return 0;
    let removed = 0;
    for (const id of disabledIds) {
      const item = snapshot.consumers.find((c) => c.id === id);
      if (item) removed += item.tokens;
    }
    return snapshot.total_tokens - removed;
  })());

  // projectedPct: projected tokens as a percentage of the context window.
  const projectedPct = $derived(
    snapshot && snapshot.window_tokens > 0
      ? (projectedTokens / snapshot.window_tokens) * 100
      : 0
  );

  // projectedDelta: how many tokens we save by disabling the selected rows.
  const projectedDelta = $derived(snapshot ? snapshot.total_tokens - projectedTokens : 0);

  // gaugeColor: severity colour for the % gauge based on thresholds in the brief.
  function gaugeColor(p: number): string {
    if (p >= 75) return 'var(--red)';
    if (p >= 50) return 'var(--amber)';
    return 'var(--accent)';
  }

  // toggleConsumer: adds or removes a consumer id from the disabled set.
  // Svelte 5 requires reassigning a $state Set to trigger reactivity.
  function toggleConsumer(id: number, checked: boolean) {
    const next = new Set(disabledIds);
    if (checked) {
      // checkbox checked → consumer is ENABLED → remove from disabled set
      next.delete(id);
    } else {
      // checkbox unchecked → consumer is DISABLED → add to disabled set
      next.add(id);
    }
    disabledIds = next;
  }

  // categoryColor: a consistent accent colour for each category in the stacked bar.
  // Cycles through a small palette — sufficient for the 4–5 categories in practice.
  const CATEGORY_COLORS = [
    'var(--accent)',
    'var(--amber)',
    'var(--green)',
    'var(--red)',
    'var(--text-dim)',
  ];
  function categoryColor(index: number): string {
    return CATEGORY_COLORS[index % CATEGORY_COLORS.length];
  }

  // categoryChipStyle: inline style for the category label chip in the table.
  // Maps known category names to their status-palette colours (same logic as
  // ScopeLadderRow's chipStyle, re-applied to category names).
  function categoryChipStyle(cat: string): string {
    const base =
      'font-size:11px;padding:2px 7px;border-radius:5px;font-weight:600;white-space:nowrap;';
    switch (cat) {
      case 'skill':
        return base + 'background:var(--accent-soft);color:var(--accent);';
      case 'mcp':
        return base + 'background:var(--amber-soft);color:var(--amber);';
      case 'hook':
        return base + 'background:var(--green-soft);color:var(--green);';
      case 'agent':
        return base + 'background:var(--red-soft);color:var(--red);';
      default:
        return base + 'background:var(--panel-2);color:var(--text-faint);';
    }
  }
</script>

<!-- ── Page header ── -->
<div class="head">
  <h1>Context Budget</h1>
  <p>The startup context tax — tokens your harness consumes before you even prompt.</p>
</div>

<!-- ── Loading / error states ── -->
{#if err}
  <div class="error">
    Couldn't load context budget: {err}
    <button onclick={load}>Retry</button>
  </div>
{:else if !snapshot}
  <p class="loading">Loading context budget…</p>

<!-- ── Empty state: total_tokens === 0 ── -->
{:else if snapshot.total_tokens === 0}
  <div class="empty-state">
    <div class="empty-icon">∑</div>
    <div class="empty-title">No always-on context detected</div>
    <div class="empty-body">
      No harness components with a non-zero context footprint were found.
      Try running a prompt and refreshing, or check that the Claude Code daemon
      is resolving your configuration.
    </div>
  </div>

<!-- ── Loaded state ── -->
{:else}

  <!-- ── Header KPI: total tokens + % gauge ── -->
  <div class="kpi-row">
    <div class="kpi-card">
      <div class="kpi-label">Always-on context tax</div>
      <div class="kpi-value">{snapshot.total_tokens.toLocaleString()} tokens</div>
      <div class="kpi-sub">{tokensCompact(snapshot.window_tokens)} token window · {pct(snapshot.pct)} used</div>
    </div>

    <!-- Horizontal % gauge bar -->
    <div class="gauge-card">
      <div class="gauge-label">
        <span>Context window usage</span>
        <span class="gauge-pct" style="color:{gaugeColor(snapshot.pct)};">
          {pct(snapshot.pct)}
        </span>
      </div>
      <div class="gauge-track">
        <div
          class="gauge-fill"
          style="width:{Math.min(100, snapshot.pct)}%;background:{gaugeColor(snapshot.pct)};"
        ></div>
      </div>
      <div class="gauge-legend">
        <span>0</span>
        <span>{tokensCompact(snapshot.window_tokens)}</span>
      </div>
    </div>
  </div>

  <!-- ── Stacked bar by category ── -->
  <div class="panel section">
    <div class="panel-head">By category</div>
    <div class="stacked-bar-wrap">
      <!-- The bar: one colored segment per category, width = segment.pct% -->
      <div class="stacked-bar" role="img" aria-label="Stacked category bar">
        {#each snapshot.by_category as seg, i}
          <div
            class="bar-seg"
            style="width:{Math.min(100, seg.pct)}%;background:{categoryColor(i)};"
            title="{seg.category}: {seg.tokens.toLocaleString()} tokens ({pct(seg.pct)})"
          ></div>
        {/each}
      </div>
      <!-- Legend: one row per category -->
      <div class="bar-legend">
        {#each snapshot.by_category as seg, i}
          <div class="legend-item">
            <span class="legend-dot" style="background:{categoryColor(i)};"></span>
            <span class="legend-cat">{seg.category}</span>
            <span class="legend-tokens">{seg.tokens.toLocaleString()}</span>
            <span class="legend-count">{seg.count} item{seg.count === 1 ? '' : 's'}</span>
            <span class="legend-pct">{pct(seg.pct)}</span>
          </div>
        {/each}
      </div>
    </div>
  </div>

  <!-- ── Biggest consumers table ── -->
  <div class="panel section">
    <div class="panel-head">Biggest consumers</div>
    <table class="consumers-table">
      <thead>
        <tr>
          <th class="col-check" scope="col"><span class="sr-only">Enable</span></th>
          <th class="col-cat"  scope="col">Category</th>
          <th class="col-name" scope="col">Name</th>
          <th class="col-scope" scope="col">Scope</th>
          <th class="col-tok"  scope="col">Tokens</th>
        </tr>
      </thead>
      <tbody>
        {#each snapshot.consumers as c (c.id)}
          {@const isDisabled = disabledIds.has(c.id)}
          <tr class:row-disabled={isDisabled}>
            <td class="col-check">
              <!--
                The checkbox is labelled by the adjacent name cell.
                checked = NOT in disabledIds (disabled set tracks exclusions).
              -->
              <input
                type="checkbox"
                id="consumer-{c.id}"
                checked={!isDisabled}
                onchange={(e) => toggleConsumer(c.id, (e.target as HTMLInputElement).checked)}
                aria-label="Enable {c.name}"
              />
            </td>
            <td class="col-cat">
              <span style={categoryChipStyle(c.category)}>{c.category}</span>
            </td>
            <td class="col-name">
              <label for="consumer-{c.id}" class="consumer-name">{c.name}</label>
            </td>
            <td class="col-scope mono">{c.scope}</td>
            <td class="col-tok">
              {c.tokens.toLocaleString()}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  <!-- ── Projected total ("if disabled" recompute) ── -->
  <div class="panel section projected">
    <div class="panel-head">Projected total (if disabled)</div>
    <div class="projected-body">
      <div class="proj-row">
        <span class="proj-label">Projected tokens</span>
        <span class="proj-val">{projectedTokens.toLocaleString()}</span>
      </div>
      <div class="proj-row">
        <span class="proj-label">Savings</span>
        <span class="proj-delta" class:delta-zero={projectedDelta === 0}>
          {projectedDelta === 0 ? 'none' : `−${projectedDelta.toLocaleString()} tokens`}
        </span>
      </div>
      <div class="proj-row">
        <span class="proj-label">Projected window %</span>
        <span class="proj-pct" style="color:{gaugeColor(projectedPct)};">
          {pct(projectedPct)}
        </span>
      </div>

      <!-- Projected % mini gauge -->
      <div class="proj-gauge-track">
        <div
          class="proj-gauge-fill"
          style="width:{Math.min(100, projectedPct)}%;background:{gaugeColor(projectedPct)};"
        ></div>
      </div>

      <p class="proj-note">
        Projection is additive; it ignores the rare case where disabling a component promotes a shadowed same-name item.
      </p>
    </div>
  </div>

  <!-- ── Caveats ── -->
  {#if snapshot.caveats.length > 0}
    <div class="caveats">
      {#each snapshot.caveats as caveat}
        <p class="caveat-line">
          <span class="caveat-icon">ⓘ</span>
          {caveat}
        </p>
      {/each}
    </div>
  {/if}

{/if}

<style>
  /* ── Page header ── */
  .head { margin-bottom: 18px; }
  h1 { margin: 0; font-size: 21px; font-weight: 600; letter-spacing: -0.02em; }
  .head p { margin: 4px 0 0; font-size: 13px; color: var(--text-faint); }

  /* ── Loading / error ── */
  .loading { padding: 16px; font-size: 13px; color: var(--text-faint); }
  .error {
    padding: 12px 16px;
    font-size: 13px;
    color: var(--red);
    border: 1px solid var(--border);
    border-radius: 8px;
    margin-bottom: 14px;
  }
  .error button {
    margin-left: 8px;
    font: inherit;
    cursor: pointer;
  }

  /* ── Empty state ── */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 60px 24px;
    border: 1px solid var(--border);
    border-radius: 12px;
    background: var(--panel);
    text-align: center;
  }
  .empty-icon {
    font-size: 40px;
    color: var(--text-faint);
    line-height: 1;
  }
  .empty-title {
    font-size: 16px;
    font-weight: 600;
  }
  .empty-body {
    font-size: 13px;
    color: var(--text-dim);
    max-width: 420px;
    line-height: 1.6;
  }

  /* ── Shared panel ── */
  .panel {
    border: 1px solid var(--border);
    border-radius: 11px;
    background: var(--panel);
    overflow: hidden;
  }
  .panel-head {
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-soft);
    font-size: 12.5px;
    font-weight: 600;
  }
  .section { margin-bottom: 14px; }

  /* ── KPI row ── */
  .kpi-row {
    display: grid;
    grid-template-columns: 1fr 1.6fr;
    gap: 14px;
    margin-bottom: 14px;
  }
  .kpi-card {
    border: 1px solid var(--border);
    border-radius: 11px;
    background: var(--panel);
    padding: 16px 18px;
  }
  .kpi-label {
    font-size: 11.5px;
    color: var(--text-faint);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 6px;
  }
  .kpi-value {
    font-size: 24px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    letter-spacing: -0.02em;
  }
  .kpi-sub {
    font-size: 12px;
    color: var(--text-dim);
    margin-top: 4px;
  }

  /* ── % gauge card ── */
  .gauge-card {
    border: 1px solid var(--border);
    border-radius: 11px;
    background: var(--panel);
    padding: 16px 18px;
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 8px;
  }
  .gauge-label {
    display: flex;
    justify-content: space-between;
    font-size: 12.5px;
    color: var(--text-dim);
  }
  .gauge-pct {
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }
  .gauge-track {
    height: 8px;
    border-radius: 5px;
    background: var(--border);
    overflow: hidden;
  }
  .gauge-fill {
    height: 100%;
    border-radius: 5px;
    transition: width 0.4s ease, background 0.3s ease;
  }
  .gauge-legend {
    display: flex;
    justify-content: space-between;
    font-size: 11px;
    color: var(--text-faint);
    font-variant-numeric: tabular-nums;
  }

  /* ── Stacked bar ── */
  .stacked-bar-wrap { padding: 14px 16px; }
  .stacked-bar {
    display: flex;
    height: 20px;
    border-radius: 6px;
    overflow: hidden;
    background: var(--border);
    margin-bottom: 14px;
  }
  .bar-seg {
    height: 100%;
    min-width: 2px;
    transition: width 0.3s ease;
  }
  .bar-legend {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .legend-item {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 12.5px;
  }
  .legend-dot {
    width: 10px;
    height: 10px;
    border-radius: 3px;
    flex: none;
  }
  .legend-cat {
    flex: 1;
    color: var(--text);
    font-weight: 500;
  }
  .legend-tokens {
    font-variant-numeric: tabular-nums;
    color: var(--text-dim);
    min-width: 80px;
    text-align: right;
  }
  .legend-count {
    color: var(--text-faint);
    min-width: 60px;
    text-align: right;
  }
  .legend-pct {
    color: var(--text-faint);
    min-width: 40px;
    text-align: right;
    font-variant-numeric: tabular-nums;
  }

  /* ── Consumers table ── */
  .consumers-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12.5px;
  }
  .consumers-table thead th {
    padding: 8px 14px;
    text-align: left;
    font-size: 11px;
    font-weight: 600;
    color: var(--text-faint);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    border-bottom: 1px solid var(--border-soft);
  }
  .consumers-table tbody tr {
    border-bottom: 1px solid var(--border-soft);
    transition: background 0.12s;
  }
  .consumers-table tbody tr:last-child { border-bottom: none; }
  .consumers-table tbody tr:hover { background: var(--panel-2); }
  .consumers-table td {
    padding: 10px 14px;
    vertical-align: middle;
  }
  .row-disabled td:not(.col-check) {
    opacity: 0.45;
  }

  /* Column widths */
  .col-check { width: 36px; }
  .col-cat   { width: 100px; }
  .col-name  { }   /* flex-fill */
  .col-scope { width: 110px; color: var(--text-dim); }
  .col-tok   { width: 100px; text-align: right; font-variant-numeric: tabular-nums; font-weight: 600; }

  /* Screen-reader only utility */
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
  }

  /* Checkbox: inherit system styling, cursor pointer */
  input[type='checkbox'] { cursor: pointer; width: 15px; height: 15px; }

  /* Consumer name label (links the checkbox via for=) */
  .consumer-name { cursor: pointer; }

  .mono { font-family: 'IBM Plex Mono', monospace; font-size: 11.5px; }

  /* ── Projected panel ── */
  .projected .panel-head { border-bottom: 1px solid var(--border-soft); }
  .projected-body { padding: 14px 16px; }
  .proj-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 6px 0;
    font-size: 13px;
    border-bottom: 1px solid var(--border-soft);
  }
  .proj-row:last-of-type { border-bottom: none; }
  .proj-label { color: var(--text-dim); }
  .proj-val {
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }
  .proj-delta {
    font-weight: 600;
    color: var(--green);
    font-variant-numeric: tabular-nums;
  }
  .proj-delta.delta-zero { color: var(--text-faint); }
  .proj-pct {
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }
  .proj-gauge-track {
    height: 6px;
    border-radius: 4px;
    background: var(--border);
    overflow: hidden;
    margin: 12px 0 10px;
  }
  .proj-gauge-fill {
    height: 100%;
    border-radius: 4px;
    transition: width 0.3s ease, background 0.3s ease;
  }
  .proj-note {
    font-size: 11.5px;
    color: var(--text-faint);
    font-style: italic;
    margin: 8px 0 0;
    line-height: 1.5;
  }

  /* ── Caveats ── */
  .caveats { margin-top: 14px; }
  .caveat-line {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    font-size: 12px;
    color: var(--text-faint);
    margin: 0 0 6px;
    line-height: 1.5;
  }
  .caveat-icon {
    color: var(--text-faint);
    flex: none;
    margin-top: 1px;
  }
</style>
