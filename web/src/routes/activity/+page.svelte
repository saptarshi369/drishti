<!--
  Live Activity — counters, live event stream, sparklines, skills-triggered table.

  Reads the single `activity` store fed by the one app EventSource (spec §12.2);
  it NEVER opens its own EventSource — the layout's connect() owns the stream.

  Functional behaviour per the M2 spec:
    • 5 counter cards: session + last-24h buckets for prompts/skills/tools/blocked/errors
    • Left panel: rolling live stream with two sparklines (prompts/min, skills/min)
    • Right panel: skills-triggered table sorted by count; dead-skill footnote if any
    • Loading state: shown while `activity` store is null (first snapshot pending)

  CSS variables used here are all confirmed in app.css:
    --green, --accent, --red, --amber, hud-blink keyframe, --border, --panel, etc.
-->
<script lang="ts">
  // Import the shared stores from $lib/sse — never create a second EventSource.
  // `activity` carries the full ActivitySnapshot; `status` is 'live'|'starting'|'offline'.
  import { activity, status } from '$lib/sse';

  // Components built in Tasks 13 & 14 — all use Svelte 5 runes ($props).
  import CounterCard from '$lib/components/CounterCard.svelte';
  import LiveEventRow from '$lib/components/LiveEventRow.svelte';
  import SkillTriggerRow from '$lib/components/SkillTriggerRow.svelte';
  import Sparkline from '$lib/components/Sparkline.svelte';

  // --- Svelte 5 runes: $derived replaces Svelte 4's `$:` reactive statements ---
  // $activity is the Svelte auto-subscription syntax (still valid in runes mode).
  // snap: the current ActivitySnapshot, or null if no data has arrived yet.
  const snap = $derived($activity);

  // cs: the counters object inside the snapshot.
  // This is only accessed inside {#if snap} so null-safety is guaranteed at render time.
  // TypeScript needs the `?.` here because $derived evaluates outside the template guard.
  const cs = $derived(snap?.counters);

  // maxSkill: the highest trigger count across all skills, used to scale the bar chart.
  // Default to 1 so SkillTriggerRow never divides by zero on first render.
  const maxSkill = $derived(snap?.skills?.reduce((m, s) => Math.max(m, s.count), 1) ?? 1);

  // deadCount: number of "dead" skills (not fired in 7+ days), drives the footnote.
  const deadCount = $derived(snap?.skills?.filter((s) => s.dead).length ?? 0);
</script>

<!-- Header: title + live status dot -->
<div class="head">
  <div>
    <h1>Live Activity</h1>
    <p>Prompts fired and skills/tools triggered — from JSONL transcripts.</p>
  </div>
  <!-- Status dot: green + blinking when connected, grey otherwise -->
  <div class="live">
    <span class="dot" class:on={$status === 'live'}></span>
    {$status === 'live' ? 'watcher streaming' : $status}
  </div>
</div>

<!-- Loading state: shown until the first ActivitySnapshot arrives from the daemon -->
{#if !snap}
  <p class="empty">Waiting for the daemon…</p>
{:else}
  <!-- 5 counter cards: session (primary) + last-24h (secondary).
       cs is guaranteed non-null here because snap is truthy and snap.counters always exists. -->
  <div class="counters">
    <CounterCard label="Prompts fired"    session={cs!.session.prompts} day={cs!.last_24h.prompts} />
    <CounterCard label="Skills triggered" session={cs!.session.skills}  day={cs!.last_24h.skills}  color="var(--accent)" />
    <CounterCard label="Tool calls"       session={cs!.session.tools}   day={cs!.last_24h.tools} />
    <CounterCard label="Blocked"          session={cs!.session.blocked} day={cs!.last_24h.blocked} color="var(--red)" />
    <CounterCard label="Errors"           session={cs!.session.errors}  day={cs!.last_24h.errors}  color="var(--amber)" />
  </div>

  <!-- Two-column grid: live stream (left, wider) + skills table (right) -->
  <div class="grid">

    <!-- LEFT: Live event stream with sparkline header -->
    <div class="panel">
      <div class="panel-head">
        <span>Live stream</span>
        <!-- Two mini sparklines: prompts/min (accent) and skills/min (green) -->
        <div class="sparks">
          <span>prompts/min</span>
          <Sparkline data={snap.sparklines.prompts_per_min} color="var(--accent)" />
          <span>skills/min</span>
          <Sparkline data={snap.sparklines.skills_per_min} color="var(--green)" />
        </div>
      </div>

      <!-- Scrollable event list.
           Key expression: ts_ms + type + tool/skill name → stable identity across updates.
           Blocked events get a red tint via LiveEventRow's .blocked class. -->
      <div class="stream">
        {#each snap.recent as ev (ev.ts_ms + ev.type + (ev.tool_name ?? ev.skill_name ?? ''))}
          <LiveEventRow {ev} />
        {/each}
        {#if snap.recent.length === 0}
          <p class="empty">No events yet.</p>
        {/if}
      </div>
    </div>

    <!-- RIGHT: Skills-triggered table -->
    <div class="panel">
      <div class="panel-head">
        <span>Skills triggered</span>
        <span class="muted">sorted by count</span>
      </div>

      <!-- One row per skill — bar width proportional to maxSkill -->
      {#each snap.skills as s (s.name)}
        <SkillTriggerRow skill={s} max={maxSkill} />
      {/each}

      <!-- Dead-skill footnote: appears when at least one skill hasn't fired in 7+ days -->
      {#if deadCount > 0}
        <div class="dead-note">
          <span>⚠</span>
          {deadCount} dead skill{deadCount > 1 ? 's' : ''} never triggered — flagged for archive.
        </div>
      {/if}
    </div>

  </div>
{/if}

<style>
  /* Page header: title left, live-dot right */
  .head { display: flex; align-items: baseline; justify-content: space-between; margin-bottom: 16px; }
  h1 { margin: 0; font-size: 21px; font-weight: 600; letter-spacing: -0.02em; }
  p { margin: 4px 0 0; font-size: 13px; color: var(--text-faint); }

  /* Status indicator row */
  .live { display: flex; align-items: center; gap: 7px; font-size: 12px; color: var(--text-dim); }
  /* Dot: grey when offline/starting, green + blinking (hud-blink keyframe from app.css) when live */
  .dot { width: 7px; height: 7px; border-radius: 50%; background: var(--text-faint); }
  .dot.on { background: var(--green); animation: hud-blink 1.6s infinite; }

  /* Counter grid: 5 equal columns */
  .counters { display: grid; grid-template-columns: repeat(5, 1fr); gap: 12px; margin-bottom: 14px; }

  /* Main two-column grid: live stream is 1.4× wider than skills table */
  .grid { display: grid; grid-template-columns: 1.4fr 1fr; gap: 14px; }

  /* Card panels */
  .panel { border: 1px solid var(--border); border-radius: 11px; background: var(--panel); overflow: hidden; }
  .panel-head {
    display: flex; align-items: center; justify-content: space-between;
    padding: 13px 16px; border-bottom: 1px solid var(--border-soft);
    font-size: 13px; font-weight: 600;
  }

  /* Sparkline row in the live-stream header */
  .sparks { display: flex; gap: 10px; align-items: center; font-size: 11px; color: var(--text-faint); font-weight: 400; }

  /* Scrollable event list */
  .stream { padding: 6px 8px; max-height: 430px; overflow-y: auto; }

  /* Muted label in skills-table header */
  .muted { color: var(--text-faint); font-weight: 400; font-size: 11px; }

  /* Dead-skill footnote at the bottom of the skills panel */
  .dead-note { padding: 11px 16px; font-size: 11.5px; color: var(--text-faint); display: flex; align-items: center; gap: 7px; }
  .dead-note span { color: var(--amber); }

  /* Empty states (loading + no-events) */
  .empty { padding: 16px; font-size: 13px; color: var(--text-faint); }
</style>
