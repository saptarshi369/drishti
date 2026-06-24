<!--
  Settings — live screen replacing the P0 stub (Module 7).

  Fetches GET /api/settings on mount (mirrors skills/+page.svelte: loading /
  error+retry / loaded states). On success renders six anchored panels in a
  single scrollable page:

    1. Appearance     — theme + accent (instant live effect + persist)
    2. Agents & Roots — read-only agent info + mutable roots list + folder picker
    3. Retention      — duration windows + disk usage estimate
    4. Daemon         — port / bind_addr / throttle + restart banner
    5. Privacy+Update — posture chips + auto-check toggle + update check
    6. Config files   — context_window_tokens, skill thresholds, security rules
    7. Live-helper    — statusline suggestion (non-mutating, copy-to-clipboard)

  Each panel imports its own component from $lib/components/settings/ so this
  file stays focused on layout and initial load state.
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import { getSettings, type SettingsView } from '$lib/api';

  import AppearancePanel from '$lib/components/settings/AppearancePanel.svelte';
  import AgentsPanel     from '$lib/components/settings/AgentsPanel.svelte';
  import RetentionPanel  from '$lib/components/settings/RetentionPanel.svelte';
  import DaemonPanel     from '$lib/components/settings/DaemonPanel.svelte';
  import PrivacyPanel    from '$lib/components/settings/PrivacyPanel.svelte';
  import ConfigFilesPanel from '$lib/components/settings/ConfigFilesPanel.svelte';
  import LiveHelperPanel from '$lib/components/settings/LiveHelperPanel.svelte';

  // snap: the fetched settings view; null while loading or on error.
  let snap = $state<SettingsView | null>(null);
  let err  = $state<string | null>(null);

  async function load() {
    err  = null;
    snap = null;
    try {
      snap = await getSettings();
    } catch (e) {
      err = e instanceof Error ? e.message : 'failed to load settings';
    }
  }

  // onSaved: called by each panel after a successful save. Re-fetches the
  // current settings so the shared snap reflects server state — preventing
  // cross-panel clobber where saving panel B reverts panel A's just-saved change.
  // Each panel reads its pass-through fields from the live snap prop at save time,
  // so once snap is refreshed all subsequent saves carry current server values.
  async function onSaved() {
    try {
      snap = await getSettings();
    } catch {
      // Best-effort: if the refresh fails, snap stays at the last known value.
      // The next successful save will bring it in sync.
    }
  }

  onMount(() => { load(); });

  // Section anchors for the sticky nav.
  const sections = [
    { id: 'appearance',   label: 'Appearance'    },
    { id: 'agents',       label: 'Agents & Roots' },
    { id: 'retention',    label: 'Retention'      },
    { id: 'daemon',       label: 'Daemon'         },
    { id: 'privacy',      label: 'Privacy'        },
    { id: 'config-files', label: 'Config files'   },
    { id: 'live-helper',  label: 'Live-helper'    },
  ];
</script>

<!-- Page header -->
<div class="head">
  <h1>Settings</h1>
  <p>Drishti daemon configuration. Changes are persisted to <code>~/.drishti/config.toml</code> and picked up within ~10 s (except port/bind changes — those need a restart).</p>
</div>

{#if err}
  <!-- Error state: matches skills/+page.svelte error pattern exactly. -->
  <div class="error">
    Couldn't load settings: {err}
    <button onclick={load}>Retry</button>
  </div>

{:else if !snap}
  <p class="loading">Loading settings…</p>

{:else}
  <!-- Anchor nav bar -->
  <nav class="section-nav" aria-label="Settings sections">
    {#each sections as s (s.id)}
      <a class="nav-link" href="#{s.id}">{s.label}</a>
    {/each}
  </nav>

  <!-- Panels — each panel receives onSaved so it can trigger a snap re-sync
       after a successful save, preventing cross-panel clobber. -->
  <div class="settings-body">
    <div class="panel-card">
      <AppearancePanel snap={snap} {onSaved} />
    </div>
    <div class="panel-card">
      <AgentsPanel snap={snap} {onSaved} />
    </div>
    <div class="panel-card">
      <RetentionPanel snap={snap} {onSaved} />
    </div>
    <div class="panel-card">
      <DaemonPanel snap={snap} {onSaved} />
    </div>
    <div class="panel-card">
      <PrivacyPanel snap={snap} {onSaved} />
    </div>
    <div class="panel-card">
      <ConfigFilesPanel snap={snap} {onSaved} />
    </div>
    <div class="panel-card">
      <LiveHelperPanel />
    </div>
  </div>
{/if}

<style>
  .head h1 { margin-bottom: 0.25rem; }
  .head p  { color: var(--text-faint); max-width: 60ch; margin: 0 0 1rem; font-size: 0.9rem; }

  .loading { color: var(--text-faint); margin-top: 1rem; }

  /* Error style: red text + inline retry — matches M6 skills screen convention. */
  .error { color: var(--red); margin-top: 1rem; display: flex; align-items: center; gap: 0.75rem; }
  .error button {
    padding: 0.25rem 0.65rem; border: 1px solid var(--red); border-radius: 5px;
    background: var(--red-soft); color: var(--red); cursor: pointer; font-size: 0.85rem;
  }
  .error button:hover { background: var(--red); color: var(--bg); }

  /* Horizontal anchor nav */
  .section-nav {
    display: flex; gap: 0.25rem; flex-wrap: wrap;
    border-bottom: 1px solid var(--border-soft);
    margin-bottom: 1.5rem; padding-bottom: 0.5rem;
  }
  .nav-link {
    font-size: 0.82rem; color: var(--text-faint); text-decoration: none;
    padding: 0.2rem 0.6rem; border-radius: 5px;
    transition: color 0.15s, background 0.15s;
  }
  .nav-link:hover { color: var(--accent); background: var(--accent-soft); }

  /* Panel layout */
  .settings-body { display: flex; flex-direction: column; gap: 1rem; }
  .panel-card {
    background: var(--panel); border-radius: 12px; padding: 1.25rem;
    border: 1px solid var(--border-soft);
  }
</style>
