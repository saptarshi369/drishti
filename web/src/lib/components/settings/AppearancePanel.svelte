<!--
  AppearancePanel: theme (dark|light) + accent (default|teal|violet) selectors.

  Theme/accent decisions:
  ─────────────────────────────────────────────────────────────────────────────
  The backend stores accent as "default"|"teal"|"violet". The theme.ts module
  uses 'indigo' as its third accent option and maps it to "no data-accent attr"
  (see applyAccent). The UI never offers 'indigo' — it shows the backend's set
  {default, teal, violet}. When the backend says "default", we call
  applyAccent('indigo') so the app-wide token system stays in sync (indigo IS
  the default token set). This mapping is:
      backend "default" ↔ theme.ts 'indigo'
      backend "teal"    ↔ theme.ts 'teal'
      backend "violet"  ↔ theme.ts 'violet'
  No internal state holds 'indigo' — only theme.ts stores/uses it; here we
  always work in the backend's three-value enum. On every change we: (1) call
  applyTheme/applyAccent for the instant live effect, AND (2) call PUT
  /api/settings so the daemon's config.toml is updated.
  ─────────────────────────────────────────────────────────────────────────────
-->
<script lang="ts">
  import { applyTheme, applyAccent } from '$lib/theme';
  import { putSettings, type SettingsView, type SettingsSaveResult } from '$lib/api';

  // snap: the current full settings — read-only here, changes propagate via PUT.
  // onSaved: called after a successful save so the parent re-fetches snap,
  // ensuring subsequent panel saves read current server state (no clobber).
  let { snap, onSaved }: { snap: SettingsView; onSaved: () => void } = $props();

  // Local mutable copies of the two appearance fields for UI binding.
  let theme = $state(snap.theme as 'dark' | 'light');
  let accent = $state(snap.accent as 'default' | 'teal' | 'violet');

  // saveState: 'idle' | 'saving' | 'saved' | 'error'
  let saveState = $state<'idle' | 'saving' | 'saved' | 'error'>('idle');
  let saveMsg = $state('');

  // toThemeArg: maps backend accent string to the theme.ts Accent type.
  // 'default' → 'indigo' (the "no accent attribute" default in theme.ts).
  function toThemeArg(a: 'default' | 'teal' | 'violet'): 'indigo' | 'teal' | 'violet' {
    return a === 'default' ? 'indigo' : a;
  }

  // onThemeChange: apply instantly AND persist. On error show inline message.
  async function onThemeChange(t: 'dark' | 'light') {
    theme = t;
    applyTheme(t);
    await save();
  }

  async function onAccentChange(a: 'default' | 'teal' | 'violet') {
    accent = a;
    applyAccent(toThemeArg(a));
    await save();
  }

  // save: persists current theme+accent to the daemon. Sends the full settings
  // object (read from snap) so no other field is clobbered by the PUT.
  async function save() {
    saveState = 'saving';
    saveMsg = '';
    try {
      const r: SettingsSaveResult = await putSettings({
        port: snap.port,
        bind_addr: snap.bind_addr,
        theme,
        accent,
        active_window: snap.active_window,
        aggregate_horizon: snap.aggregate_horizon,
        throttle: snap.throttle,
        check_interval: snap.check_interval,
        context_window_tokens: snap.context_window_tokens,
        auto_check: snap.auto_check,
      });
      if (r.warnings?.length) {
        saveState = 'saved';
        saveMsg = 'Saved. Warnings: ' + r.warnings.join('; ');
      } else {
        saveState = 'saved';
        saveMsg = 'Saved.';
      }
      onSaved();
    } catch (e) {
      saveState = 'error';
      saveMsg = e instanceof Error ? e.message : 'save failed';
    }
  }
</script>

<section class="panel" id="appearance">
  <h2>Appearance</h2>

  <div class="field-group">
    <div class="field">
      <div class="field-label" role="group" aria-label="Theme">Theme</div>
      <div class="btn-row" role="radiogroup" aria-label="Theme">
        <button
          class="toggle-btn"
          class:active={theme === 'dark'}
          onclick={() => onThemeChange('dark')}
          aria-pressed={theme === 'dark'}
        >Dark</button>
        <button
          class="toggle-btn"
          class:active={theme === 'light'}
          onclick={() => onThemeChange('light')}
          aria-pressed={theme === 'light'}
        >Light</button>
      </div>
    </div>

    <div class="field">
      <div class="field-label">Accent</div>
      <div class="btn-row" role="radiogroup" aria-label="Accent">
        <button
          class="toggle-btn"
          class:active={accent === 'default'}
          onclick={() => onAccentChange('default')}
        >Default (indigo)</button>
        <button
          class="toggle-btn"
          class:active={accent === 'teal'}
          onclick={() => onAccentChange('teal')}
        >Teal</button>
        <button
          class="toggle-btn"
          class:active={accent === 'violet'}
          onclick={() => onAccentChange('violet')}
        >Violet</button>
      </div>
    </div>
  </div>

  {#if saveState === 'saved'}
    <p class="save-ok">{saveMsg}</p>
  {:else if saveState === 'error'}
    <p class="save-err">{saveMsg}</p>
  {:else if saveState === 'saving'}
    <p class="save-info">Saving…</p>
  {/if}
</section>

<style>
  .panel { margin-bottom: 2rem; }
  .panel h2 { font-size: 0.95rem; color: var(--text-dim); margin: 0 0 1rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; }
  .field-group { display: flex; flex-direction: column; gap: 1rem; }
  .field { display: flex; flex-direction: column; gap: 0.4rem; }
  .field-label { font-size: 0.85rem; color: var(--text-faint); }
  .btn-row { display: flex; gap: 0.5rem; flex-wrap: wrap; }
  .toggle-btn {
    padding: 0.35rem 0.9rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--panel-2);
    color: var(--text-dim);
    font-size: 0.9rem;
    cursor: pointer;
    transition: border-color 0.15s, color 0.15s, background 0.15s;
  }
  .toggle-btn:hover { border-color: var(--accent); color: var(--text); }
  .toggle-btn.active {
    border-color: var(--accent);
    background: var(--accent-soft);
    color: var(--accent);
    font-weight: 600;
  }
  .save-ok   { color: var(--green);      font-size: 0.85rem; margin: 0.5rem 0 0; }
  .save-err  { color: var(--red);        font-size: 0.85rem; margin: 0.5rem 0 0; }
  .save-info { color: var(--text-faint); font-size: 0.85rem; margin: 0.5rem 0 0; }
</style>
