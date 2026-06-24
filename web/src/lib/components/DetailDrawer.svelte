<!--
  DetailDrawer: right-side slide-in panel showing the full detail for a
  selected inventory row. Opened when the user clicks a ladder row or "why?".

  Design source: docs/HarnessHUD.dc.html lines 494–540.

  Contents (top to bottom):
    • Sticky header: category icon + name + close button
    • Status chip + scope tag + token-count meta tag
    • PrecedenceTrail (fetched lazily from GET /api/inventory/{id}/why on open)
    • Source path (winner_path)
    • Attrs block (description / allowed_tools / model / transport / command / etc.)

  Props:
    row      ResolvedRow | null  — the row being inspected; null = closed
    onClose  () => void          — called when the user closes the drawer

  Closing: overlay click or ✕ button both call onClose.
  The parent controls visibility by passing row=null or row=<item>.
-->
<script lang="ts">
  import { getWhy } from '$lib/api';
  import type { ResolvedRow, TrailStep } from '$lib/api';
  import PrecedenceTrail from './PrecedenceTrail.svelte';

  let {
    row,
    onClose
  }: {
    row: ResolvedRow | null;
    onClose: () => void;
  } = $props();

  // trail: fetched on demand when the drawer opens (row becomes non-null).
  // null = not yet loaded; [] = loaded but empty; TrailStep[] = loaded.
  let trail = $state<TrailStep[] | null>(null);
  let trailError = $state<string | null>(null);
  let trailLoading = $state(false);

  // Reactive effect: whenever `row` changes to a non-null value, fetch its
  // precedence trail. When row becomes null (drawer closes), reset state.
  $effect(() => {
    if (!row) {
      trail = null;
      trailError = null;
      trailLoading = false;
      return;
    }
    // New row selected — reset and fetch.
    trail = null;
    trailError = null;
    trailLoading = true;
    getWhy(row.id)
      .then((res) => {
        trail = res.trail;
        trailLoading = false;
      })
      .catch((err) => {
        trailError = err instanceof Error ? err.message : 'Failed to load trail';
        trailLoading = false;
      });
  });

  // categoryIcon: matches the ladder row icon mapping.
  function categoryIcon(cat: string): string {
    switch (cat) {
      case 'skill':  return '✦';
      case 'mcp':    return '⬡';
      case 'hook':   return '⚓';
      case 'agent':  return '◉';
      default:       return '▪';
    }
  }

  // chipStyle: status chip colours matching ScopeLadderRow.
  function chipStyle(status: string): string {
    const base = 'font-size:11.5px;padding:3px 9px;border-radius:6px;font-weight:600;';
    switch (status) {
      case 'active':
        return base + 'background:var(--green-soft);color:var(--green);';
      case 'overridden':
        return base + 'background:var(--amber-soft);color:var(--amber);';
      default:
        return base + 'background:var(--panel-2);color:var(--text-faint);';
    }
  }

  // metaTag: renders a secondary pill tag (scope / meta information).
  function metaTagStyle(): string {
    return 'font-size:11.5px;padding:3px 9px;border-radius:6px;background:var(--panel-2);border:1px solid var(--border);color:var(--text-dim);';
  }

  // attrEntries: produce the list of key→value pairs for the Attrs section,
  // filtering out empty values so the UI stays clean.
  function attrEntries(attrs: Record<string, string>): [string, string][] {
    return Object.entries(attrs ?? {}).filter(([, v]) => v !== '');
  }

  // entries: derived list of attribute entries for the current row.
  let entries = $derived(row ? attrEntries(row.attrs) : []);
</script>

{#if row}
  <!--
    Overlay: semi-transparent backdrop behind the drawer.
    Click on it closes the drawer (mirrors the mockup's closeDrawer behaviour).
  -->
  <div
    onclick={onClose}
    role="button"
    tabindex="-1"
    aria-label="Close detail drawer"
    onkeydown={(e) => e.key === 'Escape' && onClose()}
    style="position:absolute;inset:0;background:rgba(0,0,0,.32);z-index:40;"
  ></div>

  <!--
    Drawer panel: fixed width, right edge, scrollable.
    animation: hud-drawer defined in app.css (slide from right).
  -->
  <aside
    style="
      position:absolute;
      top:0;right:0;bottom:0;
      width:420px;
      z-index:41;
      background:var(--panel);
      border-left:1px solid var(--border);
      box-shadow:var(--shadow);
      overflow-y:auto;
      animation:hud-drawer .22s ease;
    "
  >
    <!-- Sticky header (scrolls with the panel top) -->
    <div style="
      display:flex;
      align-items:flex-start;
      justify-content:space-between;
      padding:18px 20px;
      border-bottom:1px solid var(--border-soft);
      position:sticky;
      top:0;
      background:var(--panel);
      z-index:1;
    ">
      <!-- Icon + name/category -->
      <div style="display:flex;align-items:center;gap:11px;">
        <span style="
          width:34px;height:34px;
          border-radius:9px;
          background:var(--accent-soft);
          display:flex;align-items:center;justify-content:center;
          font-size:16px;
          color:var(--accent);
        ">{categoryIcon(row.category)}</span>
        <div>
          <div style="font-size:15px;font-weight:600;">{row.name}</div>
          <div style="font-size:11.5px;color:var(--text-faint);">{row.category}</div>
        </div>
      </div>
      <!-- Close button -->
      <button
        onclick={onClose}
        class="close-btn"
        style="
          width:28px;height:28px;
          border:1px solid var(--border);
          border-radius:7px;
          background:transparent;
          color:var(--text-faint);
          font-size:14px;
          cursor:pointer;
        "
      >✕</button>
    </div>

    <!-- Body: status chips, trail, source, attrs -->
    <div style="padding:18px 20px;display:flex;flex-direction:column;gap:18px;">

      <!-- Status + scope + token meta chips -->
      <div style="display:flex;gap:8px;flex-wrap:wrap;align-items:center;">
        <span style={chipStyle(row.effective_status)}>
          {row.effective_status.charAt(0).toUpperCase() + row.effective_status.slice(1)}
        </span>
        {#if row.winner_scope}
          <span style={metaTagStyle()}>{row.winner_scope}</span>
        {/if}
        {#if row.est_context_tokens > 0}
          <span style={metaTagStyle()}>~{row.est_context_tokens.toLocaleString()} tokens</span>
        {/if}
      </div>

      <!-- Precedence trail section -->
      {#if trailLoading}
        <div style="font-size:12.5px;color:var(--text-faint);padding:8px 0;">
          Loading trail…
        </div>
      {:else if trailError}
        <div style="font-size:12.5px;color:var(--amber);padding:8px 0;">
          Could not load trail: {trailError}
        </div>
      {:else if trail !== null}
        <PrecedenceTrail {trail} />
      {/if}

      <!-- Source path -->
      {#if row.winner_path}
        <div>
          <div style="font-size:11px;text-transform:uppercase;letter-spacing:.05em;color:var(--text-faint);margin-bottom:7px;">
            Source
          </div>
          <div style="
            font-family:'IBM Plex Mono',monospace;
            font-size:11.5px;
            color:var(--text-dim);
            padding:9px 12px;
            border-radius:8px;
            background:var(--bg);
            border:1px solid var(--border-soft);
            word-break:break-all;
          ">{row.winner_path}</div>
        </div>
      {/if}

      <!-- Attrs: description / allowed_tools / model / transport+command / etc. -->
      {#if entries.length > 0}
        <div>
          <div style="font-size:11px;text-transform:uppercase;letter-spacing:.05em;color:var(--text-faint);margin-bottom:7px;">
            Definition · read-only
          </div>
          <div style="
            border-radius:8px;
            background:var(--bg);
            border:1px solid var(--border-soft);
            overflow:hidden;
          ">
            {#each entries as [k, v]}
              <div style="
                display:flex;
                gap:10px;
                padding:8px 12px;
                border-bottom:1px solid var(--border-soft);
                font-size:11.5px;
              ">
                <span style="
                  font-family:'IBM Plex Mono',monospace;
                  color:var(--text-faint);
                  flex:none;
                  min-width:110px;
                ">{k}</span>
                <span style="
                  color:var(--text-dim);
                  word-break:break-all;
                  font-family:'IBM Plex Mono',monospace;
                ">{v}</span>
              </div>
            {/each}
          </div>
        </div>
      {/if}

    </div>
  </aside>
{/if}

<style>
  .close-btn:hover {
    background: var(--panel-2) !important;
    color: var(--text) !important;
  }
</style>
