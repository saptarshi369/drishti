<!--
  DaemonPanel: port, bind_addr, throttle inputs. Save → PUT /api/settings.
  Shows a prominent restart-required banner when response.restart_required
  is true (port or bind_addr changed), and inline warnings for non-fatal
  advisory messages (e.g. clamping). The daemon must be restarted manually —
  Drishti cannot restart itself.
-->
<script lang="ts">
  import { putSettings, type SettingsView, type SettingsSaveResult } from '$lib/api';

  // onSaved: called after a successful save so the parent re-fetches snap.
  let { snap, onSaved }: { snap: SettingsView; onSaved: () => void } = $props();

  // Local editable copies.
  let port     = $state(String(snap.port));
  let bindAddr = $state(snap.bind_addr);
  let throttle = $state(snap.throttle);

  let saveState      = $state<'idle' | 'saving' | 'saved' | 'error'>('idle');
  let saveMsg        = $state('');
  let restartNeeded  = $state(false);
  let warnings       = $state<string[]>([]);

  async function save() {
    saveState     = 'saving';
    saveMsg       = '';
    restartNeeded = false;
    warnings      = [];
    const portNum = parseInt(port, 10);
    if (isNaN(portNum) || portNum < 1 || portNum > 65535) {
      saveState = 'error';
      saveMsg   = 'Port must be a number between 1 and 65535.';
      return;
    }
    try {
      const r: SettingsSaveResult = await putSettings({
        port:                 portNum,
        bind_addr:            bindAddr,
        theme:                snap.theme,
        accent:               snap.accent,
        active_window:        snap.active_window,
        aggregate_horizon:    snap.aggregate_horizon,
        throttle,
        check_interval:       snap.check_interval,
        context_window_tokens: snap.context_window_tokens,
        auto_check:           snap.auto_check,
      });
      saveState     = 'saved';
      saveMsg       = 'Saved.';
      restartNeeded = r.restart_required;
      warnings      = r.warnings ?? [];
      onSaved();
    } catch (e) {
      saveState = 'error';
      saveMsg   = e instanceof Error ? e.message : 'save failed';
    }
  }
</script>

<section class="panel" id="daemon">
  <h2>Daemon</h2>

  {#if restartNeeded}
    <div class="restart-banner">
      Restart Drishti to apply the new port/address. The current listener will
      continue on the old binding until restarted.
    </div>
  {/if}

  {#if warnings.length > 0}
    <div class="warn-banner">
      {#each warnings as w (w)}
        <div>⚠ {w}</div>
      {/each}
    </div>
  {/if}

  <div class="field-group">
    <div class="field">
      <label class="field-label" for="port">Port</label>
      <input id="port" class="text-input" type="number" min="1" max="65535" bind:value={port} placeholder="7777" />
      <span class="hint">TCP port the daemon listens on. Restart required when changed.</span>
    </div>
    <div class="field">
      <label class="field-label" for="bind_addr">Bind address</label>
      <input id="bind_addr" class="text-input" type="text" bind:value={bindAddr} placeholder="127.0.0.1" />
      <span class="hint">Network interface to bind. Restart required when changed.</span>
    </div>
    <div class="field">
      <label class="field-label" for="throttle">Scheduler tick (throttle)</label>
      <input id="throttle" class="text-input" type="text" bind:value={throttle} placeholder="10s" />
      <span class="hint">How often the daemon scheduler polls (e.g. 10s). Applied without restart.</span>
    </div>
  </div>

  <div class="save-row">
    <button class="save-btn" onclick={save} disabled={saveState === 'saving'}>Save</button>
    {#if saveState === 'saved'}
      <span class="save-ok">{saveMsg}</span>
    {:else if saveState === 'error'}
      <span class="save-err">{saveMsg}</span>
    {/if}
  </div>
</section>

<style>
  .panel { margin-bottom: 2rem; }
  .panel h2 { font-size: 0.95rem; color: var(--text-dim); margin: 0 0 1rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; }
  .field-group { display: flex; flex-direction: column; gap: 0.85rem; margin-bottom: 1rem; }
  .field { display: flex; flex-direction: column; gap: 0.25rem; }
  .field-label { font-size: 0.85rem; color: var(--text-faint); }
  .text-input {
    background: var(--panel-2); border: 1px solid var(--border); border-radius: 6px;
    color: var(--text); font-size: 0.9rem; padding: 0.35rem 0.65rem;
    font-family: 'IBM Plex Mono', monospace; max-width: 20ch;
    transition: border-color 0.15s;
  }
  .text-input:focus { outline: none; border-color: var(--accent); }
  .hint { font-size: 0.78rem; color: var(--text-faint); }

  .restart-banner {
    background: var(--amber-soft); border-left: 3px solid var(--amber);
    border-radius: 6px; padding: 0.65rem 0.9rem;
    color: var(--amber); font-size: 0.875rem; margin-bottom: 1rem;
  }
  .warn-banner {
    background: var(--amber-soft); border-radius: 6px; padding: 0.55rem 0.9rem;
    color: var(--amber); font-size: 0.85rem; margin-bottom: 0.75rem;
    display: flex; flex-direction: column; gap: 0.25rem;
  }
  .save-row { display: flex; align-items: center; gap: 0.75rem; }
  .save-btn { padding: 0.35rem 0.9rem; border: 1px solid var(--border); border-radius: 6px; background: var(--panel-2); color: var(--text); font-size: 0.9rem; cursor: pointer; }
  .save-btn:hover:not(:disabled) { border-color: var(--accent); color: var(--accent); }
  .save-btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .save-ok  { color: var(--green); font-size: 0.85rem; }
  .save-err { color: var(--red);   font-size: 0.85rem; }
</style>
