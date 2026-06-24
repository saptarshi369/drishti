<!--
  Inventory (+page.svelte): Harness Map — the headline screen.

  Shows Claude's resolved configuration as a scope ladder: each skill, MCP server,
  hook, or agent with User / Project / Effective columns, status chips, and a
  "why?" drawer that renders the full precedence trail.

  Design source: docs/HarnessHUD.dc.html lines 266–310 (page), 494–540 (drawer).
  API: GET /api/inventory?category=&show_disabled=0|1  → { items: ResolvedRow[] }

  Five UI states handled (spec §12.2):
    loading  — skeleton placeholder row
    empty    — CTA explaining no items were found
    error    — inline message + retry button
    (gated/stale omitted — no gated source for inventory in M1)

  SSE: subscribes to `inventoryVersion` from $lib/sse; re-fetches on bump.
  Never opens a second EventSource — the layout's connect() owns the stream.
-->
<script lang="ts">
  import { get } from 'svelte/store';
  import { page } from '$app/stores';
  import { getInventory } from '$lib/api';
  import type { ResolvedRow } from '$lib/api';
  import { inventoryVersion, rootVersion } from '$lib/sse';
  import ScopeLadderRow from '$lib/components/ScopeLadderRow.svelte';
  import DetailDrawer from '$lib/components/DetailDrawer.svelte';

  // ——— Category tab definitions ———
  // All 8 categories are fully interactive.
  type TabDef = { id: string; label: string };

  // Tab ids are the model category strings so the ?category= filter matches the
  // resolver output directly.
  const TABS: TabDef[] = [
    { id: 'skill',        label: 'Skills'        },
    { id: 'mcp',          label: 'MCP'           },
    { id: 'hook',         label: 'Hooks'         },
    { id: 'agent',        label: 'Agents'        },
    { id: 'memory',       label: 'Memory'        },
    { id: 'command',      label: 'Commands'      },
    { id: 'output-style', label: 'Output styles' },
    { id: 'plugin',       label: 'Plugins'       },
  ];

  // ——— Reactive state ———
  // Initial tab honours a ?cat= deep-link (e.g. the Overview "Active components"
  // card links each category to /inventory?cat=hook). get(page) reads the URL once
  // at mount; an unknown/absent value falls back to the Skills tab.
  const catParam = get(page).url.searchParams.get('cat');
  let activeCategory = $state<string>(
    TABS.some((t) => t.id === catParam) ? (catParam as string) : 'skill'
  );
  let showDisabled   = $state(false);
  let filterText     = $state('');

  let rows        = $state<ResolvedRow[]>([]);
  let loading     = $state(true);
  let errorMsg    = $state<string | null>(null);

  // selectedRow: the row for which the detail drawer is open (null = closed).
  let selectedRow = $state<ResolvedRow | null>(null);

  // ——— Data fetching ———
  async function fetchRows() {
    loading  = true;
    errorMsg = null;
    try {
      const res = await getInventory(activeCategory, showDisabled);
      rows = res.items ?? [];
    } catch (e) {
      errorMsg = e instanceof Error ? e.message : 'Failed to load inventory';
      rows = [];
    } finally {
      loading = false;
    }
  }

  // Re-fetch when category or show-disabled toggle changes.
  // $effect runs after every reactive dependency changes (Svelte 5).
  $effect(() => {
    // Track both; Svelte re-runs this block when either changes.
    const _cat  = activeCategory;
    const _show = showDisabled;
    fetchRows();
  });

  // Re-fetch on SSE inventory_changed signal OR a top-bar root switch (version bumps).
  $effect(() => {
    const _v = $inventoryVersion + $rootVersion; // subscribe to both
    // Only trigger after mount (versions start at 0; skip the initial 0 value).
    if (_v > 0) fetchRows();
  });

  // ——— Derived counts ———
  // Counts are computed from ALL rows (before client-side text filter) so the
  // totals reflect what's in the DB, not what's visible in the filtered list.
  let countActive     = $derived(rows.filter((r) => r.effective_status === 'active').length);
  let countOverridden = $derived(rows.filter((r) => r.effective_status === 'overridden').length);
  let countDisabled   = $derived(
    rows.filter((r) => r.effective_status === 'disabled' || r.effective_status === 'shadowed').length
  );

  // ——— Client-side filter ———
  // Filters by name substring (case-insensitive). Applied after the API fetch.
  let filteredRows = $derived(
    filterText.trim() === ''
      ? rows
      : rows.filter((r) => r.name.toLowerCase().includes(filterText.toLowerCase()))
  );

  // ——— Tab count badge ———
  // The active tab shows the count of rows for its category; other tabs show no
  // badge (we don't pre-fetch their counts).
  function tabCount(tab: TabDef): string {
    if (tab.id === activeCategory) return String(rows.length);
    return '';
  }

  // ——— Tab style helpers ———
  function tabBorderColor(tab: TabDef): string {
    if (tab.id === activeCategory) return 'var(--accent)';
    return 'transparent';
  }
  function tabFgColor(tab: TabDef): string {
    if (tab.id === activeCategory) return 'var(--text)';
    return 'var(--text-dim)';
  }
  function tabCountBg(tab: TabDef): string {
    if (tab.id === activeCategory) return 'var(--accent-soft)';
    return 'var(--panel-2)';
  }
  function tabCountFg(tab: TabDef): string {
    if (tab.id === activeCategory) return 'var(--accent)';
    return 'var(--text-faint)';
  }

  // ——— Drawer helpers ———
  function openDrawer(row: ResolvedRow) {
    selectedRow = row;
  }
  function closeDrawer() {
    selectedRow = null;
  }

  // ——— Per-category precedence footer ———
  // Each category resolves scopes in a different order; the footer text must
  // reflect the active category so users aren't misled. $derived re-computes
  // automatically whenever activeCategory changes (Svelte 5 runes).
  let precedenceText = $derived((() => {
    switch (activeCategory) {
      case 'skill':
        return 'enterprise > user > project · deny beats allow · same-name skill beats command';
      case 'agent':
        return 'enterprise > project > user';
      case 'mcp':
        return 'local > project > user · disabled/enabled via settings';
      case 'hook':
        return 'hooks from all scopes merge — every matching hook runs';
      case 'memory':
        return 'memory files from all scopes merge into context · claudeMdExcludes hides files';
      case 'command':
        return 'enterprise > user > project · a same-name skill shadows the command';
      case 'output-style':
        return 'one active style (the outputStyle setting) · others available but not in effect';
      case 'plugin':
        return 'enabled/disabled via enabledPlugins · highest scope wins';
      default:
        return '';
    }
  })());
</script>

<!-- Page title -->
<div style="margin-bottom:16px;">
  <h1 style="margin:0;font-size:21px;font-weight:600;letter-spacing:-.02em;">Harness Map</h1>
  <p style="margin:4px 0 0;font-size:13px;color:var(--text-faint);">
    What's <em style="font-style:normal;color:var(--text-dim);">active</em> — resolved across
    user → project scope, with override trails.
  </p>
</div>

<!-- Category tabs -->
<div style="display:flex;gap:3px;border-bottom:1px solid var(--border);margin-bottom:0;overflow-x:auto;">
  {#each TABS as tab}
    <button
      onclick={() => (activeCategory = tab.id)}
      style="
        font:inherit;
        font-size:13px;
        font-weight:500;
        padding:9px 14px;
        border:none;
        background:transparent;
        cursor:pointer;
        margin-bottom:-1px;
        display:flex;
        align-items:center;
        gap:7px;
        white-space:nowrap;
        border-bottom:2px solid {tabBorderColor(tab)};
        color:{tabFgColor(tab)};
      "
    >
      {tab.label}
      {#if tabCount(tab)}
        <span style="
          font-size:10.5px;
          padding:1px 6px;
          border-radius:9px;
          background:{tabCountBg(tab)};
          color:{tabCountFg(tab)};
        ">{tabCount(tab)}</span>
      {/if}
    </button>
  {/each}
</div>

<!-- Counts line + filter + show-disabled toggle -->
<div style="display:flex;align-items:center;gap:14px;padding:13px 2px;">
  <span style="font-size:12.5px;color:var(--text-dim);">
    <span style="color:var(--green);font-weight:600;">{countActive} active</span>
    · <span style="color:var(--amber);">{countOverridden} overridden</span>
    · <span style="color:var(--text-faint);">{countDisabled} disabled</span>
  </span>
  <span style="flex:1;"></span>

  <!-- Client-side text filter -->
  <div style="
    display:flex;align-items:center;gap:7px;height:30px;
    padding:0 11px;border:1px solid var(--border);
    border-radius:7px;background:var(--panel);
    color:var(--text-faint);font-size:12.5px;
  ">
    🔎
    <input
      type="text"
      placeholder="filter…"
      bind:value={filterText}
      style="
        border:none;background:transparent;
        color:var(--text);font:inherit;font-size:12.5px;
        outline:none;width:140px;
      "
    />
  </div>

  <!-- Show-disabled toggle: uses a real <input type="checkbox"> hidden off-screen
       so a <label> can wrap it properly (a11y requirement). The visual track
       and knob are pure CSS siblings, not a button, so no nesting issues. -->
  <label style="display:flex;align-items:center;gap:7px;font-size:12.5px;color:var(--text-dim);cursor:pointer;">
    <input
      type="checkbox"
      bind:checked={showDisabled}
      style="position:absolute;opacity:0;width:0;height:0;"
    />
    <!-- Visual toggle track -->
    <span style="
      width:30px;height:17px;border-radius:10px;
      background:{showDisabled ? 'var(--accent)' : 'var(--border)'};
      position:relative;cursor:pointer;transition:.15s;flex:none;
    ">
      <span style="
        position:absolute;
        top:2px;
        left:{showDisabled ? '15px' : '2px'};
        width:13px;height:13px;
        border-radius:50%;
        background:{showDisabled ? 'white' : 'var(--text-faint)'};
        transition:.15s;
      "></span>
    </span>
    show disabled
  </label>
</div>

<!-- Ladder header row -->
<div style="
  display:grid;
  grid-template-columns:1.5fr 1fr 1fr 1.3fr;
  gap:14px;
  padding:0 16px 9px;
  font-size:11px;
  text-transform:uppercase;
  letter-spacing:.05em;
  color:var(--text-faint);
">
  <span>Component</span>
  <span>
    User <span style="font-family:'IBM Plex Mono',monospace;text-transform:none;letter-spacing:0;">~/.claude</span>
  </span>
  <span>
    Project <span style="font-family:'IBM Plex Mono',monospace;text-transform:none;letter-spacing:0;">.claude</span>
  </span>
  <span>Effective</span>
</div>

<!-- Ladder rows card -->
<div style="border:1px solid var(--border);border-radius:11px;background:var(--panel);overflow:hidden;">

  <!-- Loading state: skeleton row -->
  {#if loading}
    {#each [1, 2, 3] as _}
      <div style="
        display:grid;
        grid-template-columns:1.5fr 1fr 1fr 1.3fr;
        gap:14px;
        padding:13px 16px;
        border-bottom:1px solid var(--border-soft);
        align-items:center;
      ">
        <span style="height:12px;border-radius:4px;background:var(--panel-2);width:60%;display:block;"></span>
        <span style="height:12px;border-radius:4px;background:var(--panel-2);width:40%;display:block;"></span>
        <span style="height:12px;border-radius:4px;background:var(--panel-2);width:40%;display:block;"></span>
        <span style="height:12px;border-radius:4px;background:var(--panel-2);width:50%;display:block;"></span>
      </div>
    {/each}

  <!-- Error state -->
  {:else if errorMsg}
    <div style="padding:28px 20px;text-align:center;">
      <div style="font-size:13.5px;color:var(--text-dim);margin-bottom:10px;">
        Could not load inventory: {errorMsg}
      </div>
      <button
        onclick={fetchRows}
        style="
          font:inherit;font-size:12.5px;
          padding:7px 16px;
          border:1px solid var(--border);
          border-radius:7px;
          background:var(--panel-2);
          color:var(--text);
          cursor:pointer;
        "
      >Retry</button>
    </div>

  <!-- Empty state -->
  {:else if filteredRows.length === 0}
    <div style="padding:28px 20px;text-align:center;color:var(--text-faint);font-size:13px;">
      {#if filterText.trim()}
        No {activeCategory}s match "{filterText}".
        <button
          onclick={() => (filterText = '')}
          style="
            margin-left:8px;font:inherit;font-size:12.5px;
            padding:3px 10px;border:1px solid var(--border);
            border-radius:6px;background:var(--panel-2);
            color:var(--text-dim);cursor:pointer;
          "
        >Clear filter</button>
      {:else}
        No {activeCategory}s found.
        {#if !showDisabled}
          <button
            onclick={() => (showDisabled = true)}
            style="
              margin-left:8px;font:inherit;font-size:12.5px;
              padding:3px 10px;border:1px solid var(--border);
              border-radius:6px;background:var(--panel-2);
              color:var(--text-dim);cursor:pointer;
            "
          >Show disabled</button>
        {/if}
      {/if}
    </div>

  <!-- Populated state: one ScopeLadderRow per item -->
  {:else}
    {#each filteredRows as row (row.id)}
      <ScopeLadderRow
        {row}
        onClick={() => openDrawer(row)}
        onWhy={() => openDrawer(row)}
      />
    {/each}
  {/if}
</div>

<!-- Footer note (precedence rules summary — reactive to active category) -->
{#if precedenceText}
  <p style="margin:11px 4px 0;font-size:11.5px;color:var(--text-faint);">
    Precedence applied:
    <span style="font-family:'IBM Plex Mono',monospace;">{precedenceText}</span>
  </p>
{/if}

<!--
  DetailDrawer: rendered inside the <main> element's position:relative container
  (set by AppShell). The overlay + panel use position:absolute so they sit
  within the scrollable main area, not over the top-bar/nav.
-->
<DetailDrawer row={selectedRow} onClose={closeDrawer} />
