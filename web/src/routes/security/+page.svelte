<!--
  Security & Audit — config-security findings screen.

  Fetches GET /api/security on mount (same pattern as Context Budget / Usage).
  The API returns a SecuritySnapshot: a list of Finding items flagged by the
  security scanner, pre-sorted high→medium→low, plus severity counts and a total.

  States: loading (snapshot not yet arrived), error (fetch failed + retry),
          empty (total === 0 → "All clear"), loaded (findings list).

  Each finding is rendered as a card with a coloured left border keyed to severity:
    high   → red    (#e5484d)
    medium → amber  (#f5a623)
    low    → grey   (#8b8b8b)
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import { getSecurity, type SecuritySnapshot } from '$lib/api';
  import { rootVersion } from '$lib/sse';

  // snap: the API result, null until the mount fetch resolves.
  let snap = $state<SecuritySnapshot | null>(null);
  // err: inline error string shown with a retry button on failure.
  let err = $state<string | null>(null);

  // load: (re-)fetches the snapshot from the API, clearing the previous error.
  async function load() {
    err = null;
    try {
      snap = await getSecurity();
    } catch (e) {
      err = e instanceof Error ? e.message : 'failed to load security findings';
    }
  }

  onMount(() => { load(); });

  // Re-fetch when the top-bar root selector switches the active root (bump > 0).
  $effect(() => {
    const _v = $rootVersion;
    if (_v > 0) load();
  });

  // colour: returns a CSS colour string for a given severity level.
  // Used for the left border on each finding card and the count chip borders.
  function colour(sev: string): string {
    if (sev === 'high')   return '#e5484d';
    if (sev === 'medium') return '#f5a623';
    return '#8b8b8b';
  }

  // rank: numeric priority for each severity level (lower = shown first).
  // Unknown severities fall back to 9 so they appear at the end.
  const rank: Record<string, number> = { high: 0, medium: 1, low: 2 };

  // sortedFindings: a severity-sorted copy of snap.findings (high → medium → low).
  // Sorting client-side makes the display robust to API ordering changes.
  // We spread into a new array so we never mutate the original snap object.
  const sortedFindings = $derived(
    snap ? [...snap.findings].sort((a, b) => (rank[a.severity] ?? 9) - (rank[b.severity] ?? 9)) : []
  );
</script>

<!-- ── Page header ── -->
<div class="head">
  <h1>Security &amp; Audit</h1>
  <p>Config-security findings: missing deny rules, broad allow patterns, and other risks — each with remediation.</p>
</div>

<!-- ── Loading / error states ── -->
{#if err}
  <div class="error">
    Couldn't load security findings: {err}
    <button onclick={load}>Retry</button>
  </div>
{:else if !snap}
  <p class="loading">Loading security findings…</p>

<!-- ── Empty state: no findings ── -->
{:else if snap.total === 0}
  <div class="empty-state">
    <div class="empty-icon">✓</div>
    <div class="empty-title">All clear — no findings</div>
    <div class="empty-body">
      No security issues were detected in your current configuration.
      Drishti checks for missing deny rules, overly broad allow patterns,
      bypassPermissions, and other risks.
    </div>
  </div>

<!-- ── Loaded state ── -->
{:else}

  <!-- Severity count chips -->
  <div class="counts section">
    {#each ['high', 'medium', 'low'] as sev}
      {#if snap.counts[sev]}
        <span class="chip" style="border-color:{colour(sev)};color:{colour(sev)};">
          {snap.counts[sev]} {sev}
        </span>
      {/if}
    {/each}
  </div>

  <!-- Findings list: API pre-sorts high→medium→low -->
  <div class="panel section">
    <div class="panel-head">
      Findings
      <span class="panel-count">{snap.total} total</span>
    </div>
    <ul class="findings">
      {#each sortedFindings as f (f.rule_id + ':' + f.target_key)}
        <li class="finding" style="border-left-color:{colour(f.severity)};">
          <!-- Severity badge + title row -->
          <div class="finding-header">
            <span class="sev-badge" style="color:{colour(f.severity)};">{f.severity}</span>
            <span class="finding-title">{f.title}</span>
          </div>
          <!-- Target key: the config key or path that triggered this finding -->
          <div class="finding-target mono">{f.target_key}</div>
          <!-- Human-readable detail about why this is a risk -->
          <div class="finding-detail">{f.detail}</div>
          <!-- Remediation: what to do to fix it -->
          <div class="finding-remediation">
            <span class="rem-label">Fix:</span> {f.remediation}
          </div>
          <!-- Scope: which settings layer the finding came from -->
          {#if f.scope}
            <div class="finding-scope mono">{f.scope}</div>
          {/if}
        </li>
      {/each}
    </ul>
  </div>

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

  /* ── Section spacing ── */
  .section { margin-bottom: 14px; }

  /* ── Severity count chips ── */
  .counts {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }
  .chip {
    border: 1px solid;
    border-radius: 999px;
    padding: 2px 10px;
    font-size: 12px;
    font-weight: 600;
    letter-spacing: 0.02em;
  }

  /* ── Shared panel ── */
  .panel {
    border: 1px solid var(--border);
    border-radius: 11px;
    background: var(--panel);
    overflow: hidden;
  }
  .panel-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-soft);
    font-size: 12.5px;
    font-weight: 600;
  }
  .panel-count {
    font-size: 11.5px;
    font-weight: 400;
    color: var(--text-faint);
  }

  /* ── Findings list ── */
  .findings {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .finding {
    border-left: 4px solid;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-soft);
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .findings li:last-child { border-bottom: none; }
  .findings li:hover { background: var(--panel-2); }

  /* Finding header: severity badge + title */
  .finding-header {
    display: flex;
    align-items: baseline;
    gap: 8px;
  }
  .sev-badge {
    font-size: 10.5px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    flex: none;
  }
  .finding-title {
    font-size: 13.5px;
    font-weight: 600;
    color: var(--text);
  }

  /* Target key: the config path that triggered the finding */
  .finding-target {
    font-size: 11.5px;
    color: var(--text-dim);
    margin-top: 1px;
  }

  /* Detail: human-readable risk description */
  .finding-detail {
    font-size: 13px;
    color: var(--text);
    line-height: 1.5;
    margin-top: 4px;
  }

  /* Remediation: how to fix */
  .finding-remediation {
    font-size: 12.5px;
    color: var(--text-dim);
    line-height: 1.5;
  }
  .rem-label {
    font-weight: 600;
    color: var(--text);
  }

  /* Scope: settings layer */
  .finding-scope {
    font-size: 11px;
    color: var(--text-faint);
    margin-top: 2px;
  }

  /* Monospace for paths / keys / scope */
  .mono {
    font-family: 'IBM Plex Mono', monospace;
  }
</style>
