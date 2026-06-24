<!--
  Stub: centered placeholder page shown for routes not yet built.

  Props:
    icon        — emoji/symbol shown in the icon box (e.g. '∑')
    title       — page heading (e.g. 'Context Budget Analyzer')
    description — one-sentence description of what the page will do
    tag         — roadmap label shown in the tag pill (e.g. 'P1 · derived killer feature')

  Design source: docs/HarnessHUD.dc.html lines 482–489.
-->
<script lang="ts">
  // Props are typed to make misuse obvious at compile time.
  let { icon, title, description, tag }: {
    icon: string;
    title: string;
    description: string;
    tag: string;
  } = $props();
</script>

<!-- Full-height centred layout mirrors the mockup's flex column. -->
<div style="display:flex;flex-direction:column;align-items:center;justify-content:center;min-height:520px;text-align:center;">

  <!-- Icon box: accent-coloured symbol in a rounded panel card -->
  <div style="width:60px;height:60px;border-radius:15px;border:1px solid var(--border);background:var(--panel);display:flex;align-items:center;justify-content:center;font-size:28px;color:var(--accent);margin-bottom:20px;">
    {icon}
  </div>

  <!-- Page title -->
  <h1 style="margin:0 0 8px;font-size:22px;font-weight:600;letter-spacing:-.02em;">{title}</h1>

  <!-- Description -->
  <p style="margin:0 0 4px;max-width:440px;font-size:13.5px;color:var(--text-dim);line-height:1.6;">{description}</p>

  <!-- Roadmap tag pill with accent dot -->
  <div style="margin-top:18px;display:inline-flex;align-items:center;gap:8px;padding:6px 13px;border-radius:20px;border:1px solid var(--border);background:var(--panel);font-size:12px;color:var(--text-faint);">
    <span style="width:6px;height:6px;border-radius:50%;background:var(--accent);"></span>
    {tag}
  </div>
</div>
