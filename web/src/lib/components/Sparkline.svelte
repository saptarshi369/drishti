<!--
  Sparkline.svelte — a tiny area+line sparkline for per-minute rate arrays
  (e.g. prompts_per_min from ActivitySnapshot).

  Rendered as inline SVG, NOT canvas, on purpose: SVG resolves CSS custom
  properties (var(--accent), var(--green)) and follows the live theme, whereas a
  <canvas> stroke set to "var(--accent)" is an invalid colour the browser ignores
  — which left the line invisible. SVG also degrades gracefully when every value
  is 0 (a flat baseline near the bottom) instead of a degenerate empty chart.

  Props:
    data   — number[] of rate values
    color  — CSS colour for the line + area (default: the app accent)
    width  — viewBox width in px (default 64)
    height — viewBox height in px (default 16)
-->
<script lang="ts">
  let {
    data = [] as number[],
    color = 'var(--accent)',
    width = 64,
    height = 16,
  }: { data?: number[]; color?: string; width?: number; height?: number } = $props();

  // pad keeps the peak off the very top edge and the baseline off the very bottom.
  const pad = 2;

  // geom derives the polyline (line) and polygon (filled area) point strings from
  // the data. The y-axis is scaled to the max value (min 1 so an all-zero series
  // sits flat on the baseline rather than dividing by zero). x is spread evenly.
  const geom = $derived.by(() => {
    const n = data.length;
    if (n === 0) return { line: '', area: '' };
    const max = Math.max(...data, 1);
    const stepX = n > 1 ? width / (n - 1) : width;
    const yOf = (v: number) => height - pad - (v / max) * (height - pad * 2);
    const pts = data.map((v, i) => `${(i * stepX).toFixed(1)},${yOf(v).toFixed(1)}`);
    const line = pts.join(' ');
    // Close the area down to the bottom edge at both ends for the fill polygon.
    const area = `0,${height} ${line} ${(width).toFixed(1)},${height}`;
    return { line, area };
  });
</script>

<svg viewBox="0 0 {width} {height}" preserveAspectRatio="none" style="width:100%;height:{height}px;display:block;">
  <!-- Filled area under the line: the same colour at low opacity for a soft band. -->
  <polygon points={geom.area} style="fill:{color};opacity:0.14;" />
  <!-- The line itself. -->
  <polyline
    points={geom.line}
    fill="none"
    stroke={color}
    stroke-width="1.6"
    stroke-linejoin="round"
    stroke-linecap="round"
    vector-effect="non-scaling-stroke"
  />
</svg>
