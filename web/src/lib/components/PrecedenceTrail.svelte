<!--
  PrecedenceTrail: renders a numbered list of TrailStep items as the
  "override trail / why this resolved" block shown in the detail drawer.

  Design source: docs/HarnessHUD.dc.html lines 515–525.

  Each step shows:
    step   — monospaced step number on the left
    scope  — included in the reason text
    reason — human-readable explanation
    decision — a coloured tag chip (found/wins = green, overridden/disabled/
               shadowed = amber/faint)

  Props:
    trail  TrailStep[]  — array of precedence steps from GET /api/inventory/{id}/why
-->
<script lang="ts">
  import type { TrailStep } from '$lib/api';

  let { trail }: { trail: TrailStep[] } = $props();

  // decisionTagStyle: chip style per decision value.
  // wins/found → green (positive), overridden → amber, disabled/shadowed → faint.
  function decisionTagStyle(decision: string): string {
    const base = 'font-size:10.5px;padding:2px 7px;border-radius:5px;font-weight:600;white-space:nowrap;';
    switch (decision) {
      case 'wins':
      case 'found':
        return base + 'background:var(--green-soft);color:var(--green);';
      case 'overridden':
        return base + 'background:var(--amber-soft);color:var(--amber);';
      default: // disabled | shadowed | other
        return base + 'background:var(--panel-2);color:var(--text-faint);';
    }
  }

  // stepText: combines scope + reason into readable prose for the middle column.
  function stepText(t: TrailStep): string {
    return t.scope ? `[${t.scope}] ${t.reason}` : t.reason;
  }
</script>

<!-- Container matches design: border-radius card, amber header bar -->
<div style="border:1px solid var(--border);border-radius:10px;overflow:hidden;">
  <!-- Header strip: amber accent to signal "something was resolved / overridden" -->
  <div style="
    padding:10px 14px;
    background:var(--amber-soft);
    border-bottom:1px solid var(--border-soft);
    font-size:12px;
    font-weight:600;
    color:var(--amber);
    display:flex;
    align-items:center;
    gap:7px;
  ">
    ⤣ Override trail · why this resolved
  </div>

  <!-- Trail steps: one row per step -->
  {#each trail as t (t.step)}
    <div style="
      display:flex;
      align-items:center;
      gap:11px;
      padding:10px 14px;
      border-bottom:1px solid var(--border-soft);
    ">
      <!-- Step number: monospaced, faint, fixed width -->
      <span style="
        font-family:'IBM Plex Mono',monospace;
        font-size:11px;
        color:var(--text-faint);
        width:18px;
        text-align:center;
        flex:none;
      ">{t.step}</span>

      <!-- Scope + reason text: grows to fill available space -->
      <span style="font-size:12.5px;color:var(--text);flex:1;">{stepText(t)}</span>

      <!-- Decision tag chip -->
      <span style={decisionTagStyle(t.decision)}>{t.decision}</span>
    </div>
  {/each}

  <!-- Edge case: empty trail -->
  {#if trail.length === 0}
    <div style="padding:12px 14px;font-size:12.5px;color:var(--text-faint);">
      No trail steps recorded.
    </div>
  {/if}
</div>
