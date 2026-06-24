<!-- Overview (Command Center): KPI row + live activity + health + active-here + alerts. -->
<script lang="ts">
  import { overview, quota, activity } from '$lib/sse';
  import { usd, int, ago, pct, tokensCompact } from '$lib/format';
  import Sparkline from '$lib/components/Sparkline.svelte';
  import LiveEventRow from '$lib/components/LiveEventRow.svelte';

  // Map an alert/health severity to a Claude Design colour token.
  const sevColor: Record<string, string> = {
    red: 'var(--red)', amber: 'var(--amber)', grey: 'var(--text-faint)',
  };
  // Health bar colour: green when healthy, amber mid, red low.
  const barColor = (s: number) => (s >= 75 ? 'var(--green)' : s >= 50 ? 'var(--amber)' : 'var(--red)');
  // Map an alert/active-here CTA route key to a URL.
  const ctaHref: Record<string, string> = {
    usage: '/usage', activity: '/activity', security: '/security',
    skills: '/skills', inventory: '/inventory', context: '/context',
  };

  // Ring geometry for the health composite (r=33 → circumference ≈ 207.3).
  const RING = 2 * Math.PI * 33;
  $: health = $overview?.health;
  $: ringOffset = health ? RING * (1 - health.score / 100) : RING;
  $: comps = $overview?.active_components;
  // Headline = the largest active category; rest become the breakdown line.
  $: headline = comps?.by_category?.[0];
  $: breakdown = comps?.by_category?.slice(1) ?? [];
  $: q = $quota;
</script>

<section class="page">
  <header class="head">
    <div>
      <h1>Command Center</h1>
      <p class="sub">Everything active, live, and what it costs.</p>
    </div>
  </header>

  <!-- KPI ROW -->
  <div class="kpis">
    <a class="card kpi" href="/inventory">
      <div class="label">Active components</div>
      {#if comps}
        <div class="big">{int(headline?.count ?? 0)} <span class="unit">{headline?.category ?? ''}</span></div>
        {#if breakdown.length > 0}
        <div class="meta">
          {#each breakdown as c}<span>{c.count} {c.category}</span>{/each}
        </div>
        {/if}
      {:else}<div class="big gated">—</div>{/if}
    </a>

    <a class="card kpi" href="/activity">
      <div class="label">Prompts today</div>
      <div class="big">{$overview ? int($overview.kpis.prompts_today) : '—'}</div>
    </a>

    <a class="card kpi" href="/usage">
      <div class="label">Spend today <span class="est">est.</span></div>
      <div class="big">{$overview ? usd($overview.kpis.spend_today_usd) : '—'}</div>
    </a>

    <a class="card kpi" href="/usage">
      <div class="label">Plan quota</div>
      {#if q?.available}
        <div class="qrow"><span class="qlbl">Session</span>
          <span class="bar"><span class="fill" style="width:{pct(q.five_hour?.used_percentage ?? 0)};background:var(--amber)"></span></span>
          <span class="qval">{pct(q.five_hour?.used_percentage ?? 0)}</span></div>
        <div class="qrow"><span class="qlbl">Weekly</span>
          <span class="bar"><span class="fill" style="width:{pct(q.seven_day?.used_percentage ?? 0)};background:var(--accent)"></span></span>
          <span class="qval">{pct(q.seven_day?.used_percentage ?? 0)}</span></div>
      {:else}<div class="big gated">—</div><div class="meta">install statusline helper</div>{/if}
    </a>
  </div>

  <!-- BODY -->
  <div class="body">
    <!-- live activity -->
    <div class="card">
      <div class="cardhead"><span>Live activity</span><span class="faint">prompts/min</span></div>
      {#if $activity}
        <div class="spark"><Sparkline data={$activity.sparklines.prompts_per_min} width={300} height={46} /></div>
        <div class="events">
          {#each $activity.recent.slice(0, 6) as e (e.ts_ms + e.type + e.session_id)}
            <LiveEventRow ev={e} />
          {/each}
        </div>
      {:else}<div class="empty">No activity yet — fire a prompt in Claude Code.</div>{/if}
    </div>

    <div class="rightcol">
      <!-- health -->
      <div class="card pad">
        <div class="cardhead"><span>Harness health</span><span class="faint">composite</span></div>
        {#if health}
          <div class="healthrow">
            <div class="ring">
              <svg viewBox="0 0 80 80" class="ringsvg">
                <circle cx="40" cy="40" r="33" fill="none" stroke="var(--border)" stroke-width="7" />
                <circle cx="40" cy="40" r="33" fill="none" stroke={barColor(health.score)} stroke-width="7"
                  stroke-linecap="round" stroke-dasharray={RING} stroke-dashoffset={ringOffset} />
              </svg>
              <div class="ringval"><span class="num">{health.score}</span><span class="den">/100</span></div>
            </div>
            <div class="bars">
              {#each health.bars as h}
                <div class="barwrap">
                  <div class="barlbl"><span>{h.label}</span><span>{h.score}</span></div>
                  <div class="bar"><span class="fill" style="width:{h.score}%;background:{barColor(h.score)}"></span></div>
                </div>
              {/each}
            </div>
          </div>
        {:else}<div class="empty">Loading…</div>{/if}
      </div>

      <!-- active here -->
      <div class="card">
        <div class="cardhead"><span>Active here</span><span class="faint">· resolved</span></div>
        {#each $overview?.active_here ?? [] as s}
          <a class="hererow" href={ctaHref[s.cta] ?? '/'}>
            <span class="herelabel">{s.category === 'context' ? tokensCompact(s.count) + ' context tax' : s.count + ' ' + s.category}</span>
            <span class="faint">{s.note}</span>
          </a>
        {:else}<div class="empty">Nothing active in this project yet.</div>{/each}
      </div>
    </div>
  </div>

  <!-- alerts -->
  <div class="card alerts">
    <div class="cardhead"><span>Alerts</span>
      {#if ($overview?.alerts?.length ?? 0) > 0}<span class="badge">{$overview?.alerts?.length}</span>{/if}
    </div>
    {#each $overview?.alerts ?? [] as al}
      <div class="alertrow">
        <span class="dot" style="background:{sevColor[al.severity] ?? 'var(--text-faint)'}"></span>
        <span>{al.text}</span>
        <span class="grow"></span>
        {#if al.ts_ms}<span class="faint">{ago(al.ts_ms)}</span>{/if}
        <a class="cta" href={ctaHref[al.cta] ?? '/'}>Open</a>
      </div>
    {:else}
      <div class="empty">All clear — no alerts.</div>
    {/each}
  </div>
</section>

<style>
  .page { display: flex; flex-direction: column; gap: 14px; }
  .head h1 { margin: 0; font-size: 21px; font-weight: 600; letter-spacing: -.02em; }
  .sub { margin: 4px 0 0; font-size: 13px; color: var(--text-faint); }
  .kpis { display: grid; grid-template-columns: repeat(4, 1fr); gap: 14px; }
  .card { border: 1px solid var(--border); border-radius: 11px; background: var(--panel); overflow: hidden; }
  .card.pad, .kpi { padding: 15px 16px; }
  .kpi { text-decoration: none; color: var(--text); transition: .12s; display: block; }
  .kpi:hover { border-color: var(--text-faint); transform: translateY(-1px); }
  .label { font-size: 11.5px; color: var(--text-faint); text-transform: uppercase; letter-spacing: .05em; margin-bottom: 9px; }
  .est { opacity: .7; text-transform: none; }
  .big { font-size: 27px; font-weight: 600; letter-spacing: -.02em; font-variant-numeric: tabular-nums; }
  .big .unit { font-size: 12px; color: var(--text-dim); font-weight: 400; }
  .gated { color: var(--text-faint); }
  .meta { margin-top: 8px; display: flex; gap: 12px; font-size: 12px; color: var(--text-dim); }
  .qrow { display: flex; align-items: center; gap: 9px; margin-bottom: 6px; }
  .qlbl { font-size: 11px; color: var(--text-faint); width: 46px; }
  .qval { font-size: 12px; font-weight: 600; width: 34px; text-align: right; font-variant-numeric: tabular-nums; }
  .bar { flex: 1; height: 6px; border-radius: 3px; background: var(--border); overflow: hidden; }
  .fill { display: block; height: 100%; border-radius: 3px; }
  .body { display: grid; grid-template-columns: 1.35fr 1fr; gap: 14px; }
  .rightcol { display: flex; flex-direction: column; gap: 14px; }
  .cardhead { display: flex; align-items: center; justify-content: space-between; padding: 13px 16px; border-bottom: 1px solid var(--border-soft); font-size: 13px; font-weight: 600; }
  .faint { color: var(--text-faint); font-weight: 400; }
  .spark { padding: 14px 16px 6px; }
  .events { padding: 2px 8px 10px; }
  .empty { padding: 16px; font-size: 13px; color: var(--text-faint); }
  .healthrow { display: flex; align-items: center; gap: 16px; margin-top: 13px; }
  .ring { position: relative; width: 74px; height: 74px; flex: none; }
  .ringsvg { width: 74px; height: 74px; transform: rotate(-90deg); }
  .ringval { position: absolute; inset: 0; display: flex; flex-direction: column; align-items: center; justify-content: center; }
  .ringval .num { font-size: 22px; font-weight: 600; }
  .ringval .den { font-size: 9px; color: var(--text-faint); }
  .bars { flex: 1; display: flex; flex-direction: column; gap: 7px; }
  .barwrap .barlbl { display: flex; justify-content: space-between; font-size: 11.5px; margin-bottom: 3px; color: var(--text-dim); }
  .barwrap .bar { height: 4px; }
  .hererow { display: flex; align-items: center; gap: 10px; padding: 10px 16px; border-bottom: 1px solid var(--border-soft); text-decoration: none; color: var(--text); }
  .hererow:hover { background: var(--panel-2); }
  .herelabel { flex: 1; font-size: 13px; }
  .alertrow { display: flex; align-items: center; gap: 12px; padding: 11px 16px; border-bottom: 1px solid var(--border-soft); font-size: 13px; }
  .dot { width: 7px; height: 7px; border-radius: 50%; flex: none; }
  .grow { flex: 1; }
  .badge { font-size: 11px; padding: 1px 7px; border-radius: 10px; background: var(--amber-soft); color: var(--amber); font-weight: 500; }
  .cta { font-size: 11.5px; padding: 3px 9px; border: 1px solid var(--border); border-radius: 6px; color: var(--text-dim); text-decoration: none; }
  .cta:hover { color: var(--accent); border-color: var(--accent); }
</style>
