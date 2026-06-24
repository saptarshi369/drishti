<!--
  AgentsPanel: Agents & Roots (§4 panel 2).

  Read-only "Agents" card: Claude / ~/.claude + MCP servers list.
  Roots: list current roots with remove buttons; folder-picker via GET /api/roots?path=
  which returns { home, dirs[] }. The picker starts at `home` and lets the user
  drill into subdirectories. Selecting a directory adds it to the pending list.
  Save → PUT /api/roots with the full paths[].
-->
<script lang="ts">
  import { listDirs, setRoots, type SettingsView, type DirsResult } from '$lib/api';

  // onSaved: called after a successful save so the parent re-fetches snap.
  let { snap, onSaved }: { snap: SettingsView; onSaved: () => void } = $props();

  // Mutable copy of roots for editing before Save.
  let roots = $state<string[]>([...snap.roots]);

  // Folder picker state.
  let pickerOpen = $state(false);
  let pickerPath = $state('');  // current browsing path (empty before first open)
  let pickerHome = $state('');
  let pickerDirs = $state<string[]>([]);
  let pickerErr = $state<string | null>(null);
  let pickerLoading = $state(false);

  // Save state.
  let saveState = $state<'idle' | 'saving' | 'saved' | 'error'>('idle');
  let saveMsg = $state('');

  // openPicker: initialises the picker at the home directory.
  async function openPicker() {
    pickerOpen = true;
    pickerErr = null;
    await browseTo('');
  }

  // browseTo: fetches dirs for a given path; empty string means home.
  async function browseTo(path: string) {
    pickerLoading = true;
    pickerErr = null;
    try {
      const r: DirsResult = await listDirs(path);
      pickerHome = r.home;
      pickerPath = path || r.home;
      pickerDirs = r.dirs;
    } catch (e) {
      pickerErr = e instanceof Error ? e.message : 'failed to browse';
    } finally {
      pickerLoading = false;
    }
  }

  // goUp: navigate to the parent directory (capped at home).
  function goUp() {
    if (!pickerPath || pickerPath === pickerHome) return;
    const parent = pickerPath.split('/').slice(0, -1).join('/') || '/';
    browseTo(parent);
  }

  // selectDir: adds the directory to the roots list and closes the picker.
  function selectDir(dir: string) {
    if (!roots.includes(dir)) {
      roots = [...roots, dir];
    }
    pickerOpen = false;
  }

  // removeRoot: removes a root from the pending list.
  function removeRoot(r: string) {
    roots = roots.filter((x) => x !== r);
  }

  // save: PUT /api/roots with the current roots list.
  async function save() {
    saveState = 'saving';
    saveMsg = '';
    try {
      await setRoots(roots);
      saveState = 'saved';
      saveMsg = 'Roots saved.';
      onSaved();
    } catch (e) {
      saveState = 'error';
      saveMsg = e instanceof Error ? e.message : 'save failed';
    }
  }
</script>

<section class="panel" id="agents">
  <h2>Agents &amp; Roots</h2>

  <!-- Read-only Agents card -->
  <div class="card sub-card">
    <div class="card-label">Agents</div>
    <div class="agent-row">
      <span class="chip">Claude</span>
      <span class="chip muted">~/.claude</span>
    </div>
    {#if snap.mcp_servers.length > 0}
      <div class="mcp-label">MCP Servers</div>
      <div class="mcp-list">
        {#each snap.mcp_servers as srv (srv)}
          <span class="chip muted">{srv}</span>
        {/each}
      </div>
    {:else}
      <p class="faint">No MCP servers in inventory.</p>
    {/if}
  </div>

  <!-- Roots list -->
  <div class="roots-section">
    <div class="roots-header">
      <span class="card-label">Roots</span>
      <button class="action-btn" onclick={openPicker}>+ Add folder</button>
    </div>
    {#if roots.length === 0}
      <p class="faint">No roots configured. Drishti will use ~/.claude by default.</p>
    {:else}
      <ul class="root-list">
        {#each roots as r (r)}
          <li class="root-row">
            <span class="root-path">{r}</span>
            <button class="remove-btn" onclick={() => removeRoot(r)}>Remove</button>
          </li>
        {/each}
      </ul>
    {/if}
    <div class="save-row">
      <button class="save-btn" onclick={save} disabled={saveState === 'saving'}>Save roots</button>
      {#if saveState === 'saved'}
        <span class="save-ok">{saveMsg}</span>
      {:else if saveState === 'error'}
        <span class="save-err">{saveMsg}</span>
      {/if}
    </div>
  </div>

  <!-- Folder picker modal/overlay -->
  {#if pickerOpen}
    <div class="picker-backdrop" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) pickerOpen = false; }}>
      <div class="picker-modal" role="dialog" aria-label="Add folder">
        <div class="picker-header">
          <button class="back-btn" onclick={goUp} disabled={pickerPath === pickerHome || pickerLoading}>↑ Up</button>
          <span class="picker-path">{pickerPath || 'Loading…'}</span>
          <button class="close-btn" onclick={() => (pickerOpen = false)}>✕</button>
        </div>
        {#if pickerLoading}
          <p class="faint picker-msg">Loading…</p>
        {:else if pickerErr}
          <p class="save-err picker-msg">{pickerErr}</p>
        {:else if pickerDirs.length === 0}
          <p class="faint picker-msg">No subdirectories here.</p>
          <div class="picker-add-here">
            <button class="action-btn" onclick={() => selectDir(pickerPath)}>Add this folder</button>
          </div>
        {:else}
          <ul class="dir-list">
            <!-- Offer adding the current directory itself -->
            <li class="dir-row dir-add-self">
              <button class="dir-btn dir-self-btn" onclick={() => selectDir(pickerPath)}>
                Add this folder ({pickerPath.split('/').at(-1) || '/'})
              </button>
            </li>
            {#each pickerDirs as d (d)}
              <li class="dir-row">
                <button class="dir-btn" onclick={() => browseTo(d)}>
                  <span class="dir-icon">📁</span>
                  <span class="dir-name">{d.split('/').at(-1)}</span>
                </button>
                <button class="dir-select-btn" onclick={() => selectDir(d)}>Add</button>
              </li>
            {/each}
          </ul>
        {/if}
      </div>
    </div>
  {/if}
</section>

<style>
  .panel { margin-bottom: 2rem; }
  .panel h2 { font-size: 0.95rem; color: var(--text-dim); margin: 0 0 1rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; }

  .sub-card { background: var(--panel-2); border-radius: 8px; padding: 1rem; margin-bottom: 1rem; }
  .card-label { font-size: 0.8rem; color: var(--text-faint); margin-bottom: 0.5rem; }
  .agent-row { display: flex; gap: 0.5rem; flex-wrap: wrap; }
  .mcp-label { font-size: 0.8rem; color: var(--text-faint); margin: 0.75rem 0 0.3rem; }
  .mcp-list { display: flex; gap: 0.4rem; flex-wrap: wrap; }
  .chip { padding: 0.2rem 0.6rem; border: 1px solid var(--border); border-radius: 999px; font-size: 0.85rem; }
  .chip.muted { border-color: var(--border-soft); color: var(--text-faint); }

  .roots-section { }
  .roots-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 0.5rem; }
  .root-list { list-style: none; margin: 0 0 0.75rem; padding: 0; display: flex; flex-direction: column; gap: 0.3rem; }
  .root-row { display: flex; align-items: center; justify-content: space-between; background: var(--panel-2); border-radius: 6px; padding: 0.35rem 0.7rem; }
  .root-path { font-size: 0.85rem; font-family: 'IBM Plex Mono', monospace; color: var(--text); }
  .remove-btn { background: none; border: none; color: var(--red); cursor: pointer; font-size: 0.8rem; padding: 0; }
  .remove-btn:hover { text-decoration: underline; }
  .action-btn { padding: 0.3rem 0.75rem; border: 1px solid var(--accent); border-radius: 6px; background: var(--accent-soft); color: var(--accent); font-size: 0.85rem; cursor: pointer; }
  .action-btn:hover { background: var(--accent-dim); }
  .save-row { display: flex; align-items: center; gap: 0.75rem; margin-top: 0.5rem; }
  .save-btn { padding: 0.35rem 0.9rem; border: 1px solid var(--border); border-radius: 6px; background: var(--panel-2); color: var(--text); font-size: 0.9rem; cursor: pointer; }
  .save-btn:hover:not(:disabled) { border-color: var(--accent); color: var(--accent); }
  .save-btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .save-ok  { color: var(--green); font-size: 0.85rem; }
  .save-err { color: var(--red);   font-size: 0.85rem; }
  .faint { color: var(--text-faint); font-size: 0.85rem; margin: 0.25rem 0; }

  /* Picker overlay */
  .picker-backdrop {
    position: fixed; inset: 0; background: rgba(0,0,0,.5);
    display: flex; align-items: center; justify-content: center;
    z-index: 100;
  }
  .picker-modal {
    background: var(--panel); border-radius: 10px; width: min(480px, 92vw);
    max-height: 70vh; display: flex; flex-direction: column;
    box-shadow: var(--shadow);
  }
  .picker-header { display: flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1rem; border-bottom: 1px solid var(--border); }
  .picker-path { flex: 1; font-size: 0.8rem; font-family: 'IBM Plex Mono', monospace; color: var(--text-faint); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .back-btn, .close-btn { background: none; border: 1px solid var(--border); border-radius: 5px; color: var(--text-dim); padding: 0.2rem 0.5rem; cursor: pointer; font-size: 0.85rem; }
  .back-btn:hover:not(:disabled), .close-btn:hover { border-color: var(--accent); color: var(--accent); }
  .back-btn:disabled { opacity: 0.4; cursor: default; }
  .close-btn { margin-left: auto; }
  .picker-msg { padding: 1rem; margin: 0; }
  .picker-add-here { padding: 0 1rem 1rem; }
  .dir-list { list-style: none; margin: 0; padding: 0.5rem 0; overflow-y: auto; }
  .dir-row { display: flex; align-items: center; gap: 0.5rem; padding: 0.2rem 1rem; }
  .dir-row:hover { background: var(--panel-2); }
  .dir-add-self { border-bottom: 1px solid var(--border-soft); margin-bottom: 0.25rem; padding-bottom: 0.5rem; }
  .dir-btn { flex: 1; background: none; border: none; color: var(--text); font-size: 0.9rem; text-align: left; cursor: pointer; display: flex; align-items: center; gap: 0.4rem; padding: 0.25rem 0; }
  .dir-btn:hover { color: var(--accent); }
  .dir-self-btn { color: var(--accent); font-size: 0.85rem; }
  .dir-icon { font-size: 1rem; }
  .dir-name { }
  .dir-select-btn { background: none; border: 1px solid var(--border); border-radius: 5px; color: var(--text-dim); padding: 0.15rem 0.5rem; cursor: pointer; font-size: 0.8rem; flex-shrink: 0; }
  .dir-select-btn:hover { border-color: var(--accent); color: var(--accent); }
</style>
