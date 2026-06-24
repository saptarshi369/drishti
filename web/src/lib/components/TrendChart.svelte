<!--
  TrendChart.svelte — 7-day token trend with a top-right [ Bars | Chart ] toggle.

  Props:
    days      — DailyUsage[] (oldest→newest, length 7), each with input/cache/
                output token counts + cost_usd.
    totalCost — window total USD (header figure).

  Views:
    • Bars  (default): CSS stacked columns (output → cache → input), mockup-faithful.
    • Chart: uPlot stacked bars of the same series, with a hover tooltip.

  The chosen view persists in localStorage('drishti.usage.trendView'). uPlot is
  created once and fed via setData — never re-instantiated (spec §12.2).

  --- Chart view stacking strategy ---
  uPlot does NOT auto-stack multiple bar series; by default they GROUP (split
  the slot width, rendered side-by-side). To achieve true visual stacking we:

  1. Build CUMULATIVE series from raw input/cache/output:
       L1 = input                    (accent colour)
       L2 = input + cache            (accentDim colour)
       L3 = input + cache + output   (green colour, = total)
  2. Feed uPlot as [x, L3, L2, L1] — largest first.
  3. All three series use the SAME bars() size so each occupies the full slot
     width. Because uPlot draws series in order, L3 fills 0→total (green),
     L2 overpaints 0→(input+cache) (dim, hiding the green bottom portion),
     L1 overpaints 0→input (accent, hiding the dim bottom portion).
     Net visible bands from bottom: input(accent) | cache(dim) | output(green).

  This "paint back-to-front with full-width overlapping bars" pattern is the
  canonical uPlot stacked-bars recipe (confirmed via Context7, uPlot 1.6.32
  stacked-series demo). No external stacking helper is needed.

  --- Hover tooltip ---
  The chart uses uPlot's built-in legend as a floating tooltip (pattern from
  uPlot 1.6.32 candlestick demo, confirmed via Context7). A plugin moves the
  legend element to follow the cursor and shows raw (not cumulative) per-day
  token + cost values via setCursor hook reading from `days[idx]`.
-->
<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import uPlot from 'uplot';
  import 'uplot/dist/uPlot.min.css';
  import type { DailyUsage } from '$lib/api';
  import { usd, tokensCompact } from '$lib/format';

  let { days = [] as DailyUsage[], totalCost = 0 }: { days?: DailyUsage[]; totalCost?: number } = $props();

  // view: 'bars' | 'chart', restored from localStorage on mount.
  let view = $state<'bars' | 'chart'>('bars');
  onMount(() => {
    const saved = localStorage.getItem('drishti.usage.trendView');
    if (saved === 'chart' || saved === 'bars') view = saved;
  });
  function setView(v: 'bars' | 'chart') {
    view = v;
    localStorage.setItem('drishti.usage.trendView', v);
  }

  // --- Bars view maths ---
  // maxTotal scales every bar to the tallest day; min 1 avoids divide-by-zero.
  const maxTotal = $derived(Math.max(1, ...days.map((d) => d.total_tokens)));
  const BARS_H = 128; // px of drawable bar height
  const h = (n: number) => Math.round((n / maxTotal) * BARS_H);

  // --- Chart view (uPlot) ---
  let el = $state<HTMLDivElement | undefined>(undefined);
  let chart: uPlot | null = $state(null);

  // stackedAligned: build CUMULATIVE uPlot AlignedData for stacked rendering.
  //
  // We emit 4 arrays: [x, L3, L2, L1] where:
  //   L3 (series index 1) = input + cache + output  → green (total)
  //   L2 (series index 2) = input + cache            → accentDim
  //   L1 (series index 3) = input                    → accent
  //
  // Adding largest series first means uPlot paints L3 first (green bar from
  // 0→total), then L2 overpaints (dim bar from 0→input+cache, leaving only the
  // top output portion green), then L1 overpaints (accent bar from 0→input,
  // leaving the middle cache portion dim). This is the back-to-front overlap
  // technique: each series uses the same full slot width, so they truly stack
  // rather than grouping side-by-side.
  function stackedAligned(d: DailyUsage[]): uPlot.AlignedData {
    return [
      d.map((_, i) => i),                                                         // x: day indices
      d.map((x) => x.input_tokens + x.cache_tokens + x.output_tokens),            // L3: total (green)
      d.map((x) => x.input_tokens + x.cache_tokens),                              // L2: input+cache (dim)
      d.map((x) => x.input_tokens),                                                // L1: input (accent)
    ];
  }

  // cssVar reads a design token at runtime (uPlot needs concrete colours).
  function cssVar(name: string): string {
    return getComputedStyle(document.documentElement).getPropertyValue(name).trim() || '#888';
  }

  // tooltipPlugin: floats uPlot's legend element under the cursor and fills it
  // with raw (non-cumulative) token + cost values for the hovered day.
  //
  // The plugin follows the "legendAsTooltip" pattern from the uPlot 1.6.32
  // candlestick demo (confirmed via Context7). We use setCursor (fires on every
  // mouse move) to position the element and update its inner text. The legend is
  // kept hidden until mouseenter so there's no orphaned box when not hovering.
  //
  // `getDays` is a getter closure so the plugin always reads the current `days`
  // reactive value — safe because buildChart() is called once but the closure
  // captures the live binding through the getter.
  function tooltipPlugin(getDays: () => DailyUsage[]): uPlot.Plugin {
    let legendEl: HTMLElement;
    let overEl: HTMLElement;

    return {
      hooks: {
        init(u: uPlot) {
          // Grab the legend element uPlot already rendered and style it as a
          // floating tooltip. We use "display:none" initially; show on hover.
          legendEl = u.root.querySelector('.u-legend') as HTMLElement;
          overEl = u.over;

          // Position the legend absolutely inside the chart overlay so it
          // follows the cursor without affecting page layout.
          Object.assign(legendEl.style, {
            position: 'absolute',
            pointerEvents: 'none',
            display: 'none',
            background: 'var(--panel)',
            border: '1px solid var(--border)',
            borderRadius: '7px',
            padding: '6px 10px',
            fontSize: '11px',
            lineHeight: '1.6',
            zIndex: '100',
            whiteSpace: 'nowrap',
            color: 'var(--text)',
          });

          // Hide the small colour-square markers — we'll show text only.
          legendEl.querySelectorAll('.u-marker').forEach((m) => {
            (m as HTMLElement).style.display = 'none';
          });

          overEl.style.overflow = 'visible';
          overEl.appendChild(legendEl);

          overEl.addEventListener('mouseenter', () => { legendEl.style.display = 'block'; });
          overEl.addEventListener('mouseleave', () => { legendEl.style.display = 'none'; });
        },
        setCursor(u: uPlot) {
          const { left, top, idx } = u.cursor;
          if (idx == null || left == null || top == null) return;

          // Position tooltip to the right of the cursor with a small offset.
          // Clamp so it doesn't overflow the right edge of the chart.
          const OFFSET = 12;
          const tipW = legendEl.offsetWidth || 180;
          const clampedLeft = left + OFFSET + tipW > u.width ? left - tipW - OFFSET : left + OFFSET;
          legendEl.style.transform = `translate(${clampedLeft}px, ${Math.max(0, top - 4)}px)`;

          // Read raw (non-cumulative) values from the days prop.
          const d = getDays()[idx];
          if (!d) return;

          legendEl.innerHTML =
            `<div><b>${d.label}</b></div>` +
            `<div style="color:var(--accent)">input&nbsp;&nbsp;${tokensCompact(d.input_tokens)}</div>` +
            `<div style="color:var(--accent-dim)">cache&nbsp;&nbsp;${tokensCompact(d.cache_tokens)}</div>` +
            `<div style="color:var(--green)">output ${tokensCompact(d.output_tokens)}</div>` +
            `<div style="color:var(--text-faint)">${usd(d.cost_usd)}</div>`;
        },
      },
    };
  }

  function buildChart() {
    if (!el) return;
    const accent = cssVar('--accent');
    const accentDim = cssVar('--accent-dim');
    const green = cssVar('--green');

    // One bars path builder; all three stacked series share the same slot-width
    // factor (0.6 = 60% of slot, 40% gap) so their bars perfectly overlap.
    // If bars were different sizes they would not align and the stack would
    // appear broken. size:[0.6] confirmed valid in uPlot 1.6.32 typings:
    // BarsPathBuilderOpts.size = [factor?, max?, min?].
    const barsPath = uPlot.paths.bars!({ size: [0.6] });

    // We pass a getter closure for `days` so the tooltip plugin can always read
    // the current reactive value without being re-instantiated.
    const currentDays = () => days;

    chart = new uPlot(
      {
        width: el.clientWidth || 600,
        height: 150,
        scales: { x: { time: false } },
        // legend: show is true (needed for the tooltip plugin to find .u-legend).
        // The plugin positions and styles it as a floating tooltip; it won't
        // appear as a normal static legend below the chart.
        legend: { show: true },
        cursor: { points: { show: false } },
        axes: [
          { stroke: cssVar('--text-faint'), values: (_u, vals) => vals.map((i) => days[i]?.label ?? '') },
          { stroke: cssVar('--text-faint'), values: (_u, vals) => vals.map((v) => tokensCompact(v)) },
        ],
        // Series order matters for the stacking illusion:
        //   series[1] = L3 (total, green)  — painted first, fills 0→total
        //   series[2] = L2 (input+cache, dim) — overpaints 0→(input+cache)
        //   series[3] = L1 (input, accent) — overpaints 0→input
        // Net visible bands: input (bottom), cache (middle), output (top).
        series: [
          {},
          { label: 'output', stroke: green,     fill: green,     paths: barsPath, points: { show: false } },
          { label: 'cache',  stroke: accentDim, fill: accentDim, paths: barsPath, points: { show: false } },
          { label: 'input',  stroke: accent,    fill: accent,    paths: barsPath, points: { show: false } },
        ],
        plugins: [tooltipPlugin(currentDays)],
      },
      stackedAligned(days),
      el
    );
  }

  // Rebuild on first switch to chart view (lazily, so we don't pay for uPlot
  // when the user stays on Bars). Update via setData when data changes while
  // chart is visible. The $effect fires whenever view, chart, el, or days change.
  $effect(() => {
    if (view === 'chart' && !chart && el) buildChart();
    if (chart) chart.setData(stackedAligned(days));
  });

  onDestroy(() => chart?.destroy());
</script>

<div class="card">
  <div class="head">
    <span class="title">Tokens &amp; cost · 7 days</span>
    <div class="right">
      <span class="legend"><i style="background:var(--accent)"></i>input</span>
      <span class="legend"><i style="background:var(--accent-dim)"></i>cache</span>
      <span class="legend"><i style="background:var(--green)"></i>output</span>
      <span class="total">{usd(totalCost)}</span>
      <div class="toggle">
        <button class:on={view === 'bars'} onclick={() => setView('bars')}>Bars</button>
        <button class:on={view === 'chart'} onclick={() => setView('chart')}>Chart</button>
      </div>
    </div>
  </div>

  {#if view === 'bars'}
    <div class="bars">
      {#each days as d (d.day)}
        <div class="col" title={`${d.label}: ${tokensCompact(d.total_tokens)} tok · ${usd(d.cost_usd)}`}>
          <div class="stack">
            <span style="height:{h(d.output_tokens)}px;background:var(--green)"></span>
            <span style="height:{h(d.cache_tokens)}px;background:var(--accent-dim)"></span>
            <span style="height:{h(d.input_tokens)}px;background:var(--accent)"></span>
          </div>
          <span class="xlabel">{d.label}</span>
        </div>
      {/each}
    </div>
  {:else}
    <div class="chart" bind:this={el}></div>
  {/if}
</div>

<style>
  .card { border: 1px solid var(--border); border-radius: 11px; background: var(--panel); padding: 16px 18px; margin-bottom: 14px; }
  .head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
  .title { font-size: 13px; font-weight: 600; }
  .right { display: flex; align-items: center; gap: 14px; font-size: 11.5px; color: var(--text-faint); }
  .legend { display: flex; align-items: center; gap: 5px; }
  .legend i { width: 9px; height: 9px; border-radius: 2px; display: inline-block; }
  .total { color: var(--text); font-weight: 600; font-size: 14px; font-variant-numeric: tabular-nums; }
  .toggle { display: flex; border: 1px solid var(--border); border-radius: 7px; overflow: hidden; }
  .toggle button { font: inherit; font-size: 11px; padding: 3px 9px; background: var(--panel); color: var(--text-dim); border: 0; cursor: pointer; }
  .toggle button.on { background: var(--accent); color: #fff; }
  .bars { display: flex; align-items: flex-end; gap: 10px; height: 150px; padding-bottom: 22px; }
  .col { flex: 1; display: flex; flex-direction: column; align-items: center; height: 100%; justify-content: flex-end; position: relative; }
  .stack { width: 100%; max-width: 44px; display: flex; flex-direction: column; border-radius: 4px 4px 0 0; overflow: hidden; }
  .xlabel { position: absolute; bottom: -20px; font-size: 11px; color: var(--text-faint); }
  .chart { width: 100%; height: 150px; }
</style>
