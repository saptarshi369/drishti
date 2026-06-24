<!--
  Skills Analytics — "is each skill earning its context?"

  Fetches GET /api/skills on mount (same pattern as the Security screen). The API
  returns a SkillsSnapshot: per-skill rows with triggers, always-on context cost,
  a value ratio (triggers per 1k tokens), and dead / over-triggering / disabled
  flags, plus summary counts and the total context tax.

  States: loading, error (+retry), empty (total === 0), loaded (sortable table).
  Sorting is client-side so the display is robust to API ordering.
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import { getSkills, type SkillsSnapshot, type SkillItem } from '$lib/api';

  let snap = $state<SkillsSnapshot | null>(null);
  let err = $state<string | null>(null);

  // sortKey: which column drives the client-side sort.
  let sortKey = $state<'triggers' | 'est_context_tokens' | 'value_ratio' | 'name'>('triggers');

  async function load() {
    err = null;
    try {
      snap = await getSkills();
    } catch (e) {
      err = e instanceof Error ? e.message : 'failed to load skill analytics';
    }
  }

  onMount(() => { load(); });

  // rows: a sorted copy of snap.items (we never mutate the original). Numeric
  // columns sort descending (biggest first); name sorts ascending.
  const rows = $derived.by<SkillItem[]>(() => {
    if (!snap) return [];
    const copy = [...snap.items];
    copy.sort((a, b) => {
      if (sortKey === 'name') return a.name.localeCompare(b.name);
      return (b[sortKey] as number) - (a[sortKey] as number);
    });
    return copy;
  });

  // fmtRatio: show '—' when the ratio is 0 (0-token / disabled skills) since a
  // ratio is meaningless there; otherwise 2 significant decimals.
  function fmtRatio(it: SkillItem): string {
    if (it.est_context_tokens === 0) return '—';
    return it.value_ratio.toFixed(2);
  }

  // fmtLast: relative-ish last-fired label; '—' when never fired.
  function fmtLast(ms: number): string {
    if (!ms) return '—';
    return new Date(ms).toLocaleString();
  }
</script>

<div class="head">
  <h1>Skill Analytics</h1>
  <p>Which skills earn their context. Value ratio = triggers per 1k tokens of always-on cost. Flags: dead (never fires), over-triggering (heavy but low value), disabled (turned off).</p>
</div>

{#if err}
  <div class="error">
    Couldn't load skill analytics: {err}
    <button onclick={load}>Retry</button>
  </div>
{:else if !snap}
  <p class="loading">Loading skill analytics…</p>

{:else if snap.total === 0}
  <div class="empty-state">
    <div class="empty-icon">✦</div>
    <div class="empty-title">No skills found</div>
    <div class="empty-body">No active or disabled skills were detected in your current configuration.</div>
  </div>

{:else}
  <!-- Summary header -->
  <div class="summary section">
    <span class="chip">{snap.total} skills</span>
    <span class="chip">{snap.total_context_tokens.toLocaleString()} ctx tokens</span>
    {#if snap.counts.dead}<span class="chip warn">{snap.counts.dead} dead</span>{/if}
    {#if snap.counts.over_triggering}<span class="chip warn">{snap.counts.over_triggering} over-triggering</span>{/if}
    {#if snap.counts.disabled}<span class="chip muted">{snap.counts.disabled} disabled</span>{/if}
  </div>

  <table class="skills section">
    <thead>
      <tr>
        <th><button class="sort" onclick={() => (sortKey = 'name')}>Skill</button></th>
        <th><button class="sort" onclick={() => (sortKey = 'triggers')}>Triggers</button></th>
        <th>Last fired</th>
        <th><button class="sort" onclick={() => (sortKey = 'est_context_tokens')}>Context cost</button></th>
        <th><button class="sort" onclick={() => (sortKey = 'value_ratio')}>Value ratio</button></th>
        <th>Flags</th>
      </tr>
    </thead>
    <tbody>
      {#each rows as it (it.name)}
        <tr>
          <td>{it.name}</td>
          <td>{it.triggers}</td>
          <td>{fmtLast(it.last_fired_ms)}</td>
          <td>{it.est_context_tokens.toLocaleString()}</td>
          <td>{fmtRatio(it)}</td>
          <td>
            {#if it.dead}<span class="badge warn">dead</span>{/if}
            {#if it.over_triggering}<span class="badge warn">over-triggering</span>{/if}
            {#if it.disabled}<span class="badge muted">disabled</span>{/if}
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{/if}

<style>
  .head h1 { margin-bottom: 0.25rem; }
  .head p { color: var(--text-faint); max-width: 60ch; }
  .section { margin-top: 1rem; }
  .summary { display: flex; gap: 0.5rem; flex-wrap: wrap; }
  .chip { padding: 0.2rem 0.6rem; border: 1px solid var(--border); border-radius: 999px; font-size: 0.85rem; }
  .chip.warn { border-color: var(--amber); color: var(--amber); }
  .chip.muted { border-color: var(--border); color: var(--text-faint); }
  table.skills { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
  table.skills th, table.skills td { text-align: left; padding: 0.4rem 0.6rem; border-bottom: 1px solid var(--border-soft); }
  button.sort { background: none; border: none; color: inherit; font: inherit; cursor: pointer; padding: 0; }
  .badge { padding: 0.1rem 0.45rem; border-radius: 4px; font-size: 0.75rem; margin-right: 0.3rem; }
  .badge.warn { background: var(--amber-soft); color: var(--amber); }
  .badge.muted { background: var(--panel-2); color: var(--text-faint); }
  .loading { color: var(--text-faint); margin-top: 1rem; }
  .error { color: var(--red); margin-top: 1rem; }
</style>
