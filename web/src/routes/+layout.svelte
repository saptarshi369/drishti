<!--
  Root layout: opens the single SSE connection, shows the update banner if
  an upgrade is available, applies persisted theme/accent on mount, and
  wraps every page inside the AppShell (top bar + left nav).

  SSE must be started exactly once here — child pages must NOT call connect().
-->
<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { connect, status } from '$lib/sse';
  import { theme, accent, applyTheme, applyAccent } from '$lib/theme';
  import AppShell from '$lib/components/AppShell.svelte';

  let { children } = $props();

  // update: populated from /api/update/status when an upgrade is available.
  // null means we either haven't checked yet or no update is pending.
  let update = $state<{ available: boolean; current: string; commands: string[] } | null>(null);

  onMount(() => {
    // 1. Apply the current theme/accent to the <html> element so the DOM
    //    attributes match the store values. persist=false: the initial theme may
    //    be the time-of-day default, and we must NOT freeze it as a manual choice
    //    — only an explicit toggle (AppShell/Settings) persists. accent is always
    //    an explicit value so it persists as before.
    applyTheme($theme, false);
    applyAccent($accent);

    // 2. Open the single SSE stream. connect() returns a cleanup function that
    //    closes the EventSource when the layout unmounts (e.g. during HMR).
    const close = connect();

    // 3. Check for a pending updater notification (non-blocking; silently
    //    ignored if the endpoint is not reachable).
    fetch('/api/update/status')
      .then((r) => r.json())
      .then((u) => (update = u))
      .catch(() => {});

    return close;
  });
</script>

<!-- Update banner — shown only when the daemon reports an available upgrade.
     Amber left-border follows the design-token convention (amber = warning). -->
{#if update?.available}
  <div class="card" style="margin:1rem;border-left:3px solid var(--amber)">
    Update available. To upgrade: <code>{update.commands.join(' && ')}</code>
  </div>
{/if}

<!-- AppShell provides top bar + left nav; page content is rendered as children. -->
<AppShell>
  {@render children()}
</AppShell>
