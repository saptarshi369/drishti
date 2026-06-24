<!--
  RetentionPanel: active_window, aggregate_horizon, check_interval (duration
  strings) + disk usage estimate (db_bytes + backup_bytes). Save → PUT
  /api/settings with the full current state so no other field is clobbered.

  Duration fields are shown and sent as Go duration strings (e.g. "48h0m0s").
  The server parses them and clamps active_window to [1h, 168h], emitting a
  warning in the response if clamping occurred.
-->
<script lang="ts">
  import { putSettings, type SettingsView, type SettingsSaveResult } from '$lib/api';

  // onSaved: called after a successful save so the parent re-fetches snap.
  let { snap, onSaved }: { snap: SettingsView; onSaved: () => void } = $props();

  // Local editable copies.
  let activeWindow      = $state(snap.active_window);
  let aggregateHorizon  = $state(snap.aggregate_horizon);
  let checkInterval     = $state(snap.check_interval);

  let saveState = $state<'idle' | 'saving' | 'saved' | 'error'>('idle');
  let saveMsg   = $state('');

  // fmtBytes: human-readable byte count (B / KB / MB / GB).
  function fmtBytes(n: number): string {
    if (n < 1024) return `${n} B`;
    if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
    if (n < 1024 * 1024 * 1024) return `${(n / (1024 * 1024)).toFixed(1)} MB`;
    return `${(n / (1024 * 1024 * 1024)).toFixed(2)} GB`;
  }

  const diskTotal = $derived(fmtBytes(snap.db_bytes + snap.backup_bytes));
  const diskDB    = $derived(fmtBytes(snap.db_bytes));
  const diskBak   = $derived(fmtBytes(snap.backup_bytes));

  async function save() {
    saveState = 'saving';
    saveMsg   = '';
    try {
      const r: SettingsSaveResult = await putSettings({
        port:                 snap.port,
        bind_addr:            snap.bind_addr,
        theme:                snap.theme,
        accent:               snap.accent,
        active_window:        activeWindow,
        aggregate_horizon:    aggregateHorizon,
        throttle:             snap.throttle,
        check_interval:       checkInterval,
        context_window_tokens: snap.context_window_tokens,
        auto_check:           snap.auto_check,
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

<section class="panel" id="retention">
  <h2>Retention</h2>

  <div class="field-group">
    <div class="field">
      <label class="field-label" for="active_window">Active window (lookback)</label>
      <input id="active_window" class="text-input" type="text" bind:value={activeWindow} placeholder="48h" />
      <span class="hint">Duration e.g. 1h, 48h, 72h — clamped to [1h, 168h].</span>
    </div>
    <div class="field">
      <label class="field-label" for="aggregate_horizon">Aggregate horizon</label>
      <input id="aggregate_horizon" class="text-input" type="text" bind:value={aggregateHorizon} placeholder="720h" />
      <span class="hint">How far back usage/cost stats are retained.</span>
    </div>
    <div class="field">
      <label class="field-label" for="check_interval">Update check interval</label>
      <input id="check_interval" class="text-input" type="text" bind:value={checkInterval} placeholder="24h" />
      <span class="hint">How often to poll GitHub for new Drishti versions (when auto-check is on).</span>
    </div>
  </div>

  <!-- Disk estimate (read-only) -->
  <div class="disk-card">
    <div class="disk-title">Disk usage estimate</div>
    <div class="disk-row"><span class="disk-label">Database</span><span class="disk-val">{diskDB}</span></div>
    <div class="disk-row"><span class="disk-label">Backups</span><span class="disk-val">{diskBak}</span></div>
    <div class="disk-row total"><span class="disk-label">Total</span><span class="disk-val">{diskTotal}</span></div>
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

  .disk-card { background: var(--panel-2); border-radius: 8px; padding: 0.75rem 1rem; margin-bottom: 1rem; }
  .disk-title { font-size: 0.8rem; color: var(--text-faint); margin-bottom: 0.5rem; }
  .disk-row { display: flex; gap: 1rem; font-size: 0.85rem; padding: 0.1rem 0; }
  .disk-label { color: var(--text-faint); min-width: 8ch; }
  .disk-val { font-family: 'IBM Plex Mono', monospace; }
  .disk-row.total { border-top: 1px solid var(--border-soft); margin-top: 0.3rem; padding-top: 0.3rem; font-weight: 600; }

  .save-row { display: flex; align-items: center; gap: 0.75rem; }
  .save-btn { padding: 0.35rem 0.9rem; border: 1px solid var(--border); border-radius: 6px; background: var(--panel-2); color: var(--text); font-size: 0.9rem; cursor: pointer; }
  .save-btn:hover:not(:disabled) { border-color: var(--accent); color: var(--accent); }
  .save-btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .save-ok  { color: var(--green); font-size: 0.85rem; }
  .save-err { color: var(--red);   font-size: 0.85rem; }
</style>
