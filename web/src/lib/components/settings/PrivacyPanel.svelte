<!--
  PrivacyPanel: Privacy + Update (§4 panel 5).

  Privacy posture chips are read-only, driven by scrub_locked and
  outbound_default_off flags (both are always true by design — they confirm
  the posture). auto_check toggle persisted via PUT /api/settings.
  "Check for updates" button → GET /api/update/status?check=1 → shows
  current/latest + commands[] when an update is available.
-->
<script lang="ts">
  import { checkUpdate, putSettings, type SettingsView, type UpdateStatus, type SettingsSaveResult } from '$lib/api';

  // onSaved: called after a successful save so the parent re-fetches snap.
  let { snap, onSaved }: { snap: SettingsView; onSaved: () => void } = $props();

  // auto_check persisted locally; initial value from snap.
  let autoCheck = $state(snap.auto_check);

  // Update check state.
  let updateStatus   = $state<UpdateStatus | null>(null);
  let updateLoading  = $state(false);
  let updateErr      = $state<string | null>(null);

  // Save state for auto_check toggle.
  let saveState = $state<'idle' | 'saving' | 'saved' | 'error'>('idle');
  let saveMsg   = $state('');

  async function doCheckUpdate() {
    updateLoading = true;
    updateErr     = null;
    updateStatus  = null;
    try {
      updateStatus = await checkUpdate();
    } catch (e) {
      updateErr = e instanceof Error ? e.message : 'update check failed';
    } finally {
      updateLoading = false;
    }
  }

  async function saveAutoCheck(val: boolean) {
    autoCheck = val;
    saveState = 'saving';
    saveMsg   = '';
    try {
      const r: SettingsSaveResult = await putSettings({
        port:                 snap.port,
        bind_addr:            snap.bind_addr,
        theme:                snap.theme,
        accent:               snap.accent,
        active_window:        snap.active_window,
        aggregate_horizon:    snap.aggregate_horizon,
        throttle:             snap.throttle,
        check_interval:       snap.check_interval,
        context_window_tokens: snap.context_window_tokens,
        auto_check:           val,
      });
      saveState = 'saved';
      saveMsg   = 'Saved.';
      if (r.warnings?.length) saveMsg += ' Warnings: ' + r.warnings.join('; ');
      onSaved();
    } catch (e) {
      saveState = 'error';
      saveMsg   = e instanceof Error ? e.message : 'save failed';
    }
  }
</script>

<section class="panel" id="privacy">
  <h2>Privacy + Update</h2>

  <!-- Read-only posture chips -->
  <div class="posture-row">
    {#if snap.scrub_locked}
      <span class="chip posture-green">scrub: locked on</span>
    {/if}
    <span class="chip posture-green">bind 127.0.0.1</span>
    {#if snap.outbound_default_off}
      <span class="chip posture-green">outbound off</span>
    {/if}
  </div>
  <p class="posture-note">
    Drishti scrubs secrets from all stored data (cannot be disabled). All network calls
    are opt-in only. No data leaves your machine without explicit consent.
  </p>

  <!-- auto_check toggle -->
  <div class="auto-check-row">
    <div class="toggle-label">
      <span>Auto-check for updates</span>
      <span class="hint">Periodic check of GitHub releases. The only outbound call Drishti makes.</span>
    </div>
    <button
      class="toggle-btn"
      class:on={autoCheck}
      onclick={() => saveAutoCheck(!autoCheck)}
      disabled={saveState === 'saving'}
      aria-pressed={autoCheck}
    >
      {autoCheck ? 'On' : 'Off'}
    </button>
  </div>
  {#if saveState === 'saved'}
    <p class="save-ok">{saveMsg}</p>
  {:else if saveState === 'error'}
    <p class="save-err">{saveMsg}</p>
  {/if}

  <!-- Update check button + result -->
  <div class="update-section">
    <div class="version-row">
      <span class="version-label">Running version:</span>
      <code class="version-val">{snap.version}</code>
    </div>
    <button class="check-btn" onclick={doCheckUpdate} disabled={updateLoading}>
      {updateLoading ? 'Checking…' : 'Check for updates now'}
    </button>
    {#if updateErr}
      <p class="save-err">{updateErr}</p>
    {/if}
    {#if updateStatus}
      {#if updateStatus.available}
        <div class="update-available">
          <div class="update-badge">Update available</div>
          <div class="update-versions">
            <span>{updateStatus.current}</span>
            <span class="arrow">→</span>
            <span class="new-ver">{updateStatus.latest}</span>
          </div>
          {#if updateStatus.commands.length}
            <div class="update-cmds-label">Run to upgrade:</div>
            <pre class="update-cmds">{updateStatus.commands.join('\n')}</pre>
          {/if}
        </div>
      {:else}
        <p class="up-to-date">You're up to date ({updateStatus.current}).</p>
      {/if}
    {/if}
  </div>
</section>

<style>
  .panel { margin-bottom: 2rem; }
  .panel h2 { font-size: 0.95rem; color: var(--text-dim); margin: 0 0 1rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; }

  .posture-row { display: flex; gap: 0.5rem; flex-wrap: wrap; margin-bottom: 0.5rem; }
  .chip { padding: 0.2rem 0.7rem; border-radius: 999px; font-size: 0.85rem; border: 1px solid transparent; }
  .posture-green { background: var(--green-soft); border-color: var(--green); color: var(--green); }
  .posture-note { color: var(--text-faint); font-size: 0.82rem; margin: 0 0 1rem; max-width: 55ch; }

  .auto-check-row {
    display: flex; align-items: center; justify-content: space-between; gap: 1rem;
    background: var(--panel-2); border-radius: 8px; padding: 0.65rem 0.9rem;
    margin-bottom: 0.5rem;
  }
  .toggle-label { display: flex; flex-direction: column; gap: 0.2rem; }
  .hint { font-size: 0.78rem; color: var(--text-faint); }
  .toggle-btn {
    padding: 0.3rem 0.85rem; border: 1px solid var(--border); border-radius: 6px;
    background: var(--panel-2); color: var(--text-dim); font-size: 0.9rem;
    cursor: pointer; min-width: 4rem; transition: all 0.15s;
  }
  .toggle-btn.on { background: var(--accent-soft); border-color: var(--accent); color: var(--accent); font-weight: 600; }
  .toggle-btn:disabled { opacity: 0.5; cursor: not-allowed; }

  .save-ok  { color: var(--green); font-size: 0.85rem; margin: 0.25rem 0 0; }
  .save-err { color: var(--red);   font-size: 0.85rem; margin: 0.25rem 0 0; }

  .update-section { margin-top: 1rem; }
  .version-row { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem; }
  .version-label { font-size: 0.85rem; color: var(--text-faint); }
  .version-val { font-size: 0.85rem; color: var(--text); }
  .check-btn { padding: 0.35rem 0.9rem; border: 1px solid var(--border); border-radius: 6px; background: var(--panel-2); color: var(--text); font-size: 0.9rem; cursor: pointer; }
  .check-btn:hover:not(:disabled) { border-color: var(--accent); color: var(--accent); }
  .check-btn:disabled { opacity: 0.5; cursor: not-allowed; }

  .update-available { margin-top: 0.75rem; background: var(--amber-soft); border-left: 3px solid var(--amber); border-radius: 6px; padding: 0.65rem 0.9rem; }
  .update-badge { font-size: 0.8rem; font-weight: 700; color: var(--amber); margin-bottom: 0.3rem; }
  .update-versions { display: flex; align-items: center; gap: 0.5rem; font-family: 'IBM Plex Mono', monospace; font-size: 0.9rem; margin-bottom: 0.4rem; }
  .arrow { color: var(--text-faint); }
  .new-ver { color: var(--amber); font-weight: 600; }
  .update-cmds-label { font-size: 0.78rem; color: var(--text-faint); margin-bottom: 0.2rem; }
  .update-cmds { background: var(--panel); border-radius: 5px; padding: 0.4rem 0.6rem; font-size: 0.85rem; margin: 0; white-space: pre-wrap; overflow-x: auto; }
  .up-to-date { color: var(--green); font-size: 0.875rem; margin: 0.5rem 0 0; }
</style>
