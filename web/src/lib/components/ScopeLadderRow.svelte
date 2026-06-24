<!--
  ScopeLadderRow: one row in the scope-ladder table on the Inventory screen.

  Columns (from the design):
    Component — category icon + item name
    User       — "Active" if in_user, "—" otherwise
    Project    — "Active" if in_project, "—" otherwise
    Effective  — winner_scope label + status chip + "why?" trigger

  Props:
    row    ResolvedRow  — the resolved inventory item to display
    onWhy  () => void   — called when the user clicks the "why?" badge;
                          parent opens the detail drawer for this row
    onClick() => void   — called when the row button itself is clicked;
                          same action: opens the drawer (whole-row click area)

  Status chip colours (design tokens, status-only palette):
    active    → green  (var(--green)  / var(--green-soft))
    overridden → amber (var(--amber)  / var(--amber-soft))
    disabled  → faint  (var(--text-faint) / var(--panel-2))
    shadowed  → faint  (var(--text-faint) / var(--panel-2))
-->
<script lang="ts">
  import type { ResolvedRow } from '$lib/api';

  // Props — Svelte 5 runes style.
  let {
    row,
    onWhy,
    onClick
  }: {
    row: ResolvedRow;
    onWhy: () => void;
    onClick: () => void;
  } = $props();

  // categoryIcon: a small Unicode icon for each category type.
  // Matches the design mockup's icon column intent.
  function categoryIcon(cat: string): string {
    switch (cat) {
      case 'skill':  return '✦';
      case 'mcp':    return '⬡';
      case 'hook':   return '⚓';
      case 'agent':  return '◉';
      default:       return '▪';
    }
  }

  // iconColor: active rows get the accent colour; faint otherwise.
  function iconColor(status: string): string {
    return status === 'active' ? 'var(--accent)' : 'var(--text-faint)';
  }

  // chipStyle: inline style string for the status chip based on effective_status.
  // Only green/amber/faint used — status-only palette per design rules.
  function chipStyle(status: string): string {
    const base = 'font-size:11px;padding:2px 7px;border-radius:5px;font-weight:600;white-space:nowrap;';
    switch (status) {
      case 'active':
        return base + 'background:var(--green-soft);color:var(--green);';
      case 'overridden':
        return base + 'background:var(--amber-soft);color:var(--amber);';
      default: // disabled | shadowed
        return base + 'background:var(--panel-2);color:var(--text-faint);';
    }
  }

  // chipLabel: display label for the status chip (title-cased).
  function chipLabel(status: string): string {
    return status.charAt(0).toUpperCase() + status.slice(1);
  }

  // scopePresence: "Active" if the flag is true, "—" when absent.
  // Used for the User and Project columns.
  function scopePresence(flag: boolean): string {
    return flag ? 'Active' : '—';
  }
</script>

<!--
  The row uses a <div role="row"> with a click handler. We can't nest a <button>
  inside a <button> (HTML spec violation), so the outer element is a div styled
  like a button and the "why?" badge is a real <button> inside it.
  grid-template-columns mirrors the ladder header (1.5fr 1fr 1fr 1.3fr).
-->
<div
  onclick={onClick}
  onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && onClick()}
  role="button"
  tabindex="0"
  class="ladder-row"
  style="
    width:100%;
    display:grid;
    grid-template-columns:1.5fr 1fr 1fr 1.3fr;
    gap:14px;
    align-items:center;
    padding:12px 16px;
    border-bottom:1px solid var(--border-soft);
    background:transparent;
    color:var(--text);
    cursor:pointer;
    text-align:left;
  "
>
  <!-- Component column: icon + name -->
  <span style="display:flex;align-items:center;gap:9px;min-width:0;">
    <span style="width:16px;text-align:center;font-size:13px;color:{iconColor(row.effective_status)};">
      {categoryIcon(row.category)}
    </span>
    <span style="font-weight:500;font-size:13px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">
      {row.name}
    </span>
  </span>

  <!-- User column (~/.claude presence) -->
  <span style="font-size:12px;color:var(--text-dim);">
    {scopePresence(row.in_user)}
  </span>

  <!-- Project column (.claude presence) -->
  <span style="font-size:12px;color:var(--text-dim);">
    {scopePresence(row.in_project)}
  </span>

  <!-- Effective column: winner_scope + status chip + "why?" -->
  <span style="display:flex;align-items:center;gap:8px;min-width:0;">
    <span style="font-size:12.5px;color:var(--text);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">
      {row.winner_scope || '—'}
    </span>
    <span style={chipStyle(row.effective_status)}>
      {chipLabel(row.effective_status)}
    </span>
    <!-- "why?" button: stopPropagation prevents the row-click from also firing. -->
    <button
      onclick={(e) => { e.stopPropagation(); onWhy(); }}
      class="why-btn"
      style="
        font-size:10.5px;
        padding:2px 7px;
        border-radius:5px;
        border:1px solid var(--border);
        background:var(--panel-2);
        color:var(--text-faint);
        cursor:pointer;
        white-space:nowrap;
        font:inherit;
      "
    >why?</button>
  </span>
</div>

<style>
  /* Hover highlight on the row. */
  .ladder-row:hover {
    background: var(--panel-2) !important;
  }
  .why-btn:hover {
    background: var(--panel) !important;
    color: var(--text) !important;
  }
</style>
