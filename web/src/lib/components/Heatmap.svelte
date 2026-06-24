<!--
  Heatmap.svelte — 56-day activity grid coloured by intensity bucket, with the
  current streak in the header and a less→more legend footer.

  Props:
    cells  — { day, total_tokens, bucket }[] (oldest→newest, length up to 56).
    streak — current streak in days.

  Bucket → colour: 0 border (idle), 1 accent-dim, 2/3 accent (busy).
-->
<script lang="ts">
  import { tokensCompact } from '$lib/format';

  let { cells = [] as { day: number; total_tokens: number; bucket: number }[], streak = 0 } = $props();

  // bg maps a bucket to a CSS token.
  function bg(bucket: number): string {
    if (bucket <= 0) return 'var(--border)';
    if (bucket === 1) return 'var(--accent-dim)';
    return 'var(--accent)';
  }
</script>

<div class="card">
  <div class="head">
    <span>Activity heatmap</span>
    <span class="streak">streak {streak}d</span>
  </div>
  <div class="body">
    <div class="grid">
      {#each cells as c (c.day)}
        <span class="cell" title={`${tokensCompact(c.total_tokens)} tokens`} style="background:{bg(c.bucket)}"></span>
      {/each}
    </div>
    <div class="legend">
      less
      <span class="swatches">
        <span style="background:var(--border)"></span>
        <span style="background:var(--accent-dim)"></span>
        <span style="background:var(--accent)"></span>
      </span>
      more
    </div>
  </div>
</div>

<style>
  .card { border: 1px solid var(--border); border-radius: 11px; background: var(--panel); overflow: hidden; }
  .head { padding: 12px 16px; border-bottom: 1px solid var(--border-soft); font-size: 12.5px; font-weight: 600; display: flex; justify-content: space-between; }
  .streak { color: var(--text-faint); font-weight: 400; }
  .body { padding: 14px 16px; }
  .grid { display: grid; grid-template-columns: repeat(10, 1fr); gap: 4px; }
  .cell { aspect-ratio: 1; border-radius: 2.5px; }
  .legend { display: flex; align-items: center; gap: 7px; margin-top: 12px; font-size: 11px; color: var(--text-faint); }
  .swatches { display: flex; gap: 3px; }
  .swatches span { width: 10px; height: 10px; border-radius: 2px; }
</style>
