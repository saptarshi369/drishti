<!--
  AppShell: top bar + left nav that wraps every page.

  Props:
    children  — Svelte 5 snippet; rendered as the main content area.

  Wires:
    theme / accent stores + applyTheme / applyAccent  → $lib/theme
    status store (live | starting | offline)           → $lib/sse
    page store for active-nav highlight                → $app/stores

  Disabled controls (enabled in later modules):
    • agent switcher  — M2
    • range selector  — M2
    • quota mini-gauge — M2
    • root cycler     — M2 (roots API not yet built)

  Design source: docs/HarnessHUD.dc.html lines 72–144 + navDef 696–700 + stub map 630–636.
-->
<script lang="ts">
  import { page } from '$app/stores';
  import { theme, accent, applyTheme, applyAccent } from '$lib/theme';

  // status store tells us whether the SSE daemon connection is live.
  import { status } from '$lib/sse';

  // children: the page slot in Svelte 5 runes style.
  let { children } = $props();

  // ---------- nav definition (from mockup navDef lines 696–700) ----------
  // Each entry: [href, icon, label, tag]
  // tag = '' means no badge; 'P1'/'P2' show a small pill.
  const NAV = [
    ['/',          '▦', 'Overview',  ''],
    ['/inventory', '◧', 'Inventory', ''],
    ['/activity',  '◉', 'Activity',  ''],
    ['/usage',     '◫', 'Usage',     ''],
    ['/context',   '∑', 'Context',   'P1'],
    ['/security',  '⚿', 'Security',  'P1'],
    ['/skills',    '✦', 'Skills',    'P1'],
    ['/sessions',  '⤺', 'Sessions',  'P2'],
    ['/settings',  '⚙', 'Settings',  ''],
  ] as const;

  // ---------- accent colour map for the accent picker chips ----------
  // Colours from the mockup accents array (line 725).
  const ACCENTS = [
    { id: 'indigo' as const, label: 'Indigo', col: 'oklch(0.66 0.155 264)' },
    { id: 'teal'   as const, label: 'Teal',   col: 'oklch(0.7 0.115 195)'  },
    { id: 'violet' as const, label: 'Violet', col: 'oklch(0.64 0.19 305)'  },
  ];

  // isActive: true when the current pathname matches this nav entry's href.
  // '/' must be an exact match; others are prefix-matched so sub-routes highlight
  // the correct nav item automatically.
  function isActive(href: string, pathname: string): boolean {
    if (href === '/') return pathname === '/';
    return pathname.startsWith(href);
  }

  // handleThemeToggle: flip between dark and light and persist the choice.
  function handleThemeToggle() {
    const next = $theme === 'dark' ? 'light' : 'dark';
    theme.set(next);
    applyTheme(next);
  }

  // handleAccent: set the chosen accent and update the DOM attribute.
  function handleAccent(id: 'indigo' | 'teal' | 'violet') {
    accent.set(id);
    applyAccent(id);
  }

  // themeIcon: ☀ in dark mode (click for light), ☽ in light mode (click for dark).
  // Derived reactively from the store value: handled automatically by template binding.
</script>

<!-- ============ OUTER SHELL ============ -->
<div
  style="height:100vh;display:flex;flex-direction:column;background:var(--bg);color:var(--text);font-family:'IBM Plex Sans',system-ui,sans-serif;font-size:15px;overflow:hidden;"
>

  <!-- ============ TOP BAR ============ (mockup lines 72–126) -->
  <header style="flex:none;height:52px;display:flex;align-items:center;gap:14px;padding:0 14px 0 16px;border-bottom:1px solid var(--border);background:var(--bg-2);">

    <!-- Logo / wordmark -->
    <div style="display:flex;align-items:center;gap:9px;width:204px;flex:none;">
      <div style="width:22px;height:22px;border-radius:6px;background:var(--accent);display:flex;align-items:center;justify-content:center;color:#fff;font-size:13px;font-weight:700;">▣</div>
      <span style="font-weight:600;letter-spacing:-.01em;">Drishti</span>
    </div>

    <!-- Root display — static placeholder until M2 adds a roots API.
         Shown as disabled to make the future state obvious. -->
    <button
      disabled
      aria-disabled="true"
      title="Root selection — enabled in a later module"
      style="display:flex;align-items:center;gap:7px;height:30px;padding:0 11px;border:1px solid var(--border);border-radius:7px;background:var(--panel);color:var(--text);font:inherit;font-size:12.5px;cursor:not-allowed;opacity:0.55;"
    >
      <span style="color:var(--text-faint);">⌂ Root</span>
      <span style="font-family:'IBM Plex Mono',monospace;color:var(--text);">~</span>
      <span style="color:var(--text-faint);font-size:9px;">▾</span>
    </button>

    <!-- Agent switcher — DISABLED until M2.
         Rendered with reduced opacity + aria-disabled so it is visually
         present (matches mockup) but clearly not interactive. -->
    <div
      aria-disabled="true"
      title="Agent switcher — enabled in a later module"
      style="display:flex;align-items:center;background:var(--panel);border:1px solid var(--border);border-radius:7px;height:30px;padding:2px;gap:1px;opacity:0.45;pointer-events:none;"
    >
      {#each ['Claude', 'Codex', 'Both'] as label}
        <button
          disabled
          style="font:inherit;font-size:12px;font-weight:500;height:24px;padding:0 11px;border:none;border-radius:5px;cursor:not-allowed;background:transparent;color:var(--text-dim);"
        >{label}</button>
      {/each}
    </div>

    <!-- Range selector — DISABLED until M2. -->
    <div
      aria-disabled="true"
      title="Range selector — enabled in a later module"
      style="display:flex;align-items:center;background:var(--panel);border:1px solid var(--border);border-radius:7px;height:30px;padding:2px;gap:1px;opacity:0.45;pointer-events:none;"
    >
      {#each ['Live', '24h', '7d', '30d'] as label}
        <button
          disabled
          style="font:inherit;font-size:12px;font-weight:500;height:24px;padding:0 11px;border:none;border-radius:5px;cursor:not-allowed;background:transparent;color:var(--text-dim);"
        >{label}</button>
      {/each}
    </div>

    <!-- Spacer pushes right-side controls to the far end -->
    <div style="flex:1;"></div>

    <!-- Live status dot (mockup lines 99–105).
         Colour maps to the SSE status: green = live, amber = starting, red = offline. -->
    <div style="display:flex;align-items:center;gap:7px;font-size:12.5px;color:var(--text-dim);">
      <span style="position:relative;display:inline-flex;width:8px;height:8px;">
        <span
          class="dot {$status}"
          style="position:absolute;inset:0;border-radius:50%;animation:hud-blink 1.6s ease-in-out infinite;"
        ></span>
        {#if $status === 'live'}
          <!-- Pulse ring only when live, same as mockup -->
          <span
            style="position:absolute;inset:0;border-radius:50%;background:var(--green);animation:hud-ring 1.8s ease-out infinite;"
          ></span>
        {/if}
      </span>
      <span style="color:var(--text);font-weight:500;">
        {$status === 'live' ? 'Live' : $status === 'starting' ? 'Starting…' : 'Offline'}
      </span>
    </div>

    <!-- Quota mini-gauge — DISABLED until M2 adds usage API.
         Mockup lines 107–116. Shown at reduced opacity as a visual placeholder. -->
    <button
      disabled
      aria-disabled="true"
      title="Quota gauge — enabled in a later module"
      style="display:flex;align-items:center;gap:10px;height:30px;padding:0 11px;border:1px solid var(--border);border-radius:7px;background:var(--panel);cursor:not-allowed;opacity:0.45;"
    >
      <span style="display:flex;align-items:center;gap:5px;font-size:11px;color:var(--text-faint);">⚡ Session
        <span style="display:inline-block;width:38px;height:5px;border-radius:3px;background:var(--border);overflow:hidden;">
          <span style="display:block;height:100%;border-radius:3px;background:var(--amber);width:0%;"></span>
        </span>
        <span style="color:var(--text);font-variant-numeric:tabular-nums;font-weight:600;">—</span>
      </span>
      <span style="display:flex;align-items:center;gap:5px;font-size:11px;color:var(--text-faint);">Week
        <span style="display:inline-block;width:38px;height:5px;border-radius:3px;background:var(--border);overflow:hidden;">
          <span style="display:block;height:100%;border-radius:3px;background:var(--accent);width:0%;"></span>
        </span>
        <span style="color:var(--text);font-variant-numeric:tabular-nums;font-weight:600;">—</span>
      </span>
    </button>

    <!-- Accent picker chips (mockup lines 118–123) -->
    <div style="display:flex;align-items:center;gap:3px;background:var(--panel);border:1px solid var(--border);border-radius:7px;height:30px;padding:0 6px;">
      <span style="font-size:10px;color:var(--text-faint);margin-right:2px;">Accent</span>
      {#each ACCENTS as c}
        <button
          onclick={() => handleAccent(c.id)}
          title={c.label}
          style="width:17px;height:17px;border-radius:50%;cursor:pointer;box-shadow:0 0 0 1px var(--border);padding:0;background:{c.col};border:2px solid {$accent === c.id ? 'var(--text)' : 'transparent'};"
        ></button>
      {/each}
    </div>

    <!-- Theme toggle (mockup line 125) -->
    <button
      onclick={handleThemeToggle}
      style="width:30px;height:30px;flex:none;display:flex;align-items:center;justify-content:center;border:1px solid var(--border);border-radius:7px;background:var(--panel);color:var(--text-dim);font-size:14px;cursor:pointer;"
      title="Toggle theme"
    >
      {$theme === 'dark' ? '☀' : '☽'}
    </button>
  </header>

  <!-- ============ BODY (nav + main) ============ -->
  <div style="flex:1;display:flex;min-height:0;">

    <!-- ============ LEFT NAV ============ (mockup lines 129–143) -->
    <nav style="flex:none;width:204px;border-right:1px solid var(--border);background:var(--bg-2);padding:12px;display:flex;flex-direction:column;gap:2px;overflow-y:auto;">
      {#each NAV as [href, icon, label, tag]}
        {@const active = isActive(href, $page.url.pathname)}
        <a
          {href}
          class="nav-link"
          style="display:flex;align-items:center;gap:10px;width:100%;padding:9px 11px 9px 9px;border-radius:7px;text-decoration:none;font-size:14.5px;transition:.12s;background:{active ? 'var(--accent-soft)' : 'transparent'};color:{active ? 'var(--text)' : 'var(--text-dim)'};"
        >
          <!-- Accent bar: visible only on the active route -->
          <span style="width:2px;height:15px;border-radius:2px;flex:none;background:var(--accent);transition:.12s;opacity:{active ? 1 : 0};"></span>
          <!-- Route icon -->
          <span style="width:18px;text-align:center;font-size:15px;color:{active ? 'var(--accent)' : 'var(--text-faint)'};">{icon}</span>
          <!-- Route label -->
          <span style="flex:1;">{label}</span>
          <!-- Optional priority tag (P1 / P2) -->
          {#if tag}
            <span style="font-size:9.5px;font-weight:600;padding:1px 5px;border-radius:5px;background:var(--panel-2);color:var(--text-faint);">{tag}</span>
          {/if}
        </a>
      {/each}

      <!-- Spacer + daemon info footer (mockup lines 139–143) -->
      <div style="flex:1;"></div>
      <div style="padding:10px;border:1px solid var(--border-soft);border-radius:8px;background:var(--panel);font-size:11px;color:var(--text-faint);line-height:1.5;">
        <div style="display:flex;align-items:center;gap:6px;margin-bottom:5px;color:var(--text-dim);font-weight:500;">
          <span class="dot {$status}" style="width:6px;height:6px;border-radius:50%;flex:none;"></span>
          Daemon · 127.0.0.1:7777
        </div>
        Zero outbound · local store
      </div>
    </nav>

    <!-- ============ MAIN CONTENT ============ -->
    <main style="flex:1;min-width:0;overflow-y:auto;position:relative;">
      <!-- zoom:1.1 enlarges all page content ~10% (the design's base text was too
           small to read comfortably). zoom magnifies + reflows, so fluid grids
           still fit the column width and just grow taller (main scrolls). It lives
           here, not on the 100vh shell root, so the fixed top-bar/nav stay pinned. -->
      <div style="max-width:1180px;margin:0 auto;padding:26px 30px 60px;zoom:1.1;">
        {@render children()}
      </div>
    </main>
  </div>
</div>

<style>
  .nav-link:hover { background: var(--panel) !important; }
</style>
