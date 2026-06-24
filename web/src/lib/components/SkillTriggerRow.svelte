<!--
  SkillTriggerRow.svelte — one row of the skills-triggered table.
  Props: skill (SkillStat from $lib/sse), max (largest count for bar scale).
  Dead skills (not fired in 7+ days) render in amber with a ⚠ prefix.
-->
<script lang="ts">
  // Reuse the shared SkillStat type from sse.ts (added in Task 13).
  // The brief inlined the shape; we import the named type instead to keep
  // a single source of truth aligned with the Go model.
  import type { SkillStat } from '$lib/sse';
  import { ago } from '$lib/format';

  // Props (Svelte 5 runes style — matching Sparkline.svelte's $props() idiom).
  // max defaults to 1 so we never divide by zero when the table first renders.
  let {
    skill,
    max = 1,
  }: {
    skill: SkillStat;
    max?: number;
  } = $props();

  // pct: fill percentage for the relative-width bar (0–100).
  // $derived replaces Svelte 4's `$:` reactive statement.
  // Math.max(max, 1) guards against max=0 from upstream.
  const pct = $derived(Math.round((skill.count / Math.max(max, 1)) * 100));
</script>

<div class="row">
  <span class="name" class:dead={skill.dead}>{skill.dead ? '⚠ ' : ''}{skill.name}</span>
  <span class="bar"><span class="fill" class:dead={skill.dead} style="width:{pct}%"></span></span>
  <span class="count" class:dead={skill.dead}>{skill.count}</span>
  <span class="last">{skill.count ? ago(skill.last_fired_ms) : 'never'}</span>
</div>

<style>
  .row { display: flex; align-items: center; gap: 12px; padding: 10px 16px; border-bottom: 1px solid var(--border-soft); }
  .name { font-size: 13px; color: var(--text); width: 128px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .name.dead, .count.dead { color: var(--amber); }
  .bar { flex: 1; height: 5px; border-radius: 3px; background: var(--border); overflow: hidden; }
  .fill { display: block; height: 100%; border-radius: 3px; background: var(--accent); }
  .fill.dead { background: var(--amber); }
  .count { font-size: 12.5px; font-weight: 600; font-variant-numeric: tabular-nums; width: 24px; text-align: right; color: var(--text); }
  .last { font-size: 11px; color: var(--text-faint); width: 62px; text-align: right; }
</style>
