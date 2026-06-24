<!--
  Sparkline.svelte — minimal uPlot line chart for per-minute rate arrays.

  Props:
    data   — number[] of rate values (e.g. prompts_per_min from ActivitySnapshot)
    color  — CSS colour string for the line (default: the app accent CSS var)
    width  — canvas width in px (default: 64)
    height — canvas height in px (default: 16)

  Key design constraint (spec §12.2): NEVER re-instantiate uPlot on data change.
  Instead, call chart.setData() on the live instance. Re-instantiation is
  expensive and causes a visible flash. The $effect below does exactly this.

  uPlot AlignedData shape: [[x0, x1, …], [y0, y1, …]]
    - data[0] = x-axis values (we use sequential indices since we don't have
                real timestamps for per-minute buckets)
    - data[1] = y-axis values (the rate array from the server)
-->
<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import uPlot from 'uplot';
  // uPlot ships its own CSS for the canvas container — import it here so the
  // chart element is positioned and sized correctly without extra global styles.
  import 'uplot/dist/uPlot.min.css';

  // --- Props (Svelte 5 runes style, matching the project's other components) ---
  let {
    data = [] as number[],
    color = 'var(--accent)',
    width = 64,
    height = 16,
  }: {
    data?: number[];
    color?: string;
    width?: number;
    height?: number;
  } = $props();

  // el: the <div> that uPlot will mount its canvas into (bound via bind:this).
  // Svelte 5 runes: bind:this targets must be $state so the reactivity system
  // tracks the assignment. Typed as HTMLDivElement | undefined because it starts
  // as undefined and is assigned by Svelte when the element mounts.
  let el = $state<HTMLDivElement | undefined>(undefined);

  // chart: the live uPlot instance. null until onMount runs.
  let chart: uPlot | null = $state(null);

  // toAligned: convert a flat number[] into uPlot's AlignedData format.
  // data[0] must be the x-axis array (we use sequential indices 0, 1, 2…);
  // data[1] is the y-axis array (the rate values).
  function toAligned(d: number[]): uPlot.AlignedData {
    return [d.map((_, i) => i), d];
  }

  onMount(() => {
    // Create the uPlot instance once. Options are minimal — this is a sparkline,
    // not a full chart, so we hide axes, cursor, and legend entirely.
    chart = new uPlot(
      {
        width,
        height,
        // cursor: hide the hover crosshair (we don't want interactivity on a sparkline).
        cursor: { show: false },
        // legend: hide the series legend below the chart.
        legend: { show: false },
        // scales: disable time interpretation on x so we can use plain indices.
        scales: { x: { time: false } },
        // axes: hide both x and y axes entirely (just a bare line).
        axes: [{ show: false }, { show: false }],
        // series: first entry is the required x-series placeholder; second is
        // the y-series with our colour, a thin 1.4px stroke, no dots.
        series: [
          {},
          {
            stroke: color,
            width: 1.4,
            points: { show: false },
          },
        ],
      },
      toAligned(data),
      el  // mount the chart's canvas into our <div>
    );
  });

  // Cleanup: destroy the uPlot instance when this component is removed from
  // the DOM to free canvas memory and detach any resize listeners.
  onDestroy(() => {
    chart?.destroy();
  });

  // Reactive update: when the `data` prop changes, push new values into the
  // live uPlot instance via setData. This is the correct pattern (spec §12.2):
  // - chart.setData(newData) updates the canvas in-place — no DOM teardown.
  // - We pass resetScales=true (the default) so the y-axis auto-fits each update.
  // The guard `chart &&` ensures we don't call setData before onMount creates it.
  // Note: this $effect also fires once immediately after onMount sets `chart`,
  // which is a harmless duplicate of the initial data passed to the constructor.
  // A future reader can safely ignore that first call — it's a no-op in effect.
  $effect(() => {
    if (chart) {
      chart.setData(toAligned(data));
    }
  });
</script>

<!-- The uPlot constructor appends its own <div><canvas></canvas></div> here. -->
<div bind:this={el}></div>
