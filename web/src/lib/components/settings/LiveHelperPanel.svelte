<!--
  LiveHelperPanel: statusline suggestion (§4 panel 7 / "Live-helper").

  "Generate suggestion" → GET /api/install/statusline → shows proposed in a
  <pre> + copy-to-clipboard button + added[] list + path. The copy text
  includes a prominent review note: the user must paste it into their
  settings.json themselves — Drishti never edits their file.

  Note: statusline ONLY. No hooks suggestion. Non-mutating by design (the
  handler's doc comment guarantees it).
-->
<script lang="ts">
  import { proposeStatusline, type StatuslineResult } from '$lib/api';

  let result   = $state<StatuslineResult | null>(null);
  let loading  = $state(false);
  let err      = $state<string | null>(null);
  let copied   = $state(false);

  async function generate() {
    loading = true;
    err     = null;
    result  = null;
    copied  = false;
    try {
      result = await proposeStatusline();
    } catch (e) {
      err = e instanceof Error ? e.message : 'failed to generate suggestion';
    } finally {
      loading = false;
    }
  }

  // copyReview: copies the proposed JSON plus the mandatory review note.
  async function copyReview() {
    if (!result) return;
    const text = [
      '// Review this and paste it into ~/.claude/settings.json yourself,',
      '// then verify the file is valid. Drishti never edits your file.',
      '',
      result.proposed,
    ].join('\n');
    try {
      await navigator.clipboard.writeText(text);
      copied = true;
      // Reset "Copied!" badge after 2.5 s.
      setTimeout(() => { copied = false; }, 2500);
    } catch {
      // Clipboard API can fail in non-HTTPS or restricted contexts.
    }
  }
</script>

<section class="panel" id="live-helper">
  <h2>Live-helper (statusline)</h2>

  <p class="intro">
    Generates a suggested <code>statusLine</code> entry for your Claude Code
    <code>~/.claude/settings.json</code>. Drishti never writes to your settings file —
    review the suggestion and paste it yourself.
  </p>

  <button class="gen-btn" onclick={generate} disabled={loading}>
    {loading ? 'Generating…' : 'Generate suggestion'}
  </button>

  {#if err}
    <p class="save-err">{err} <button class="retry-btn" onclick={generate}>Retry</button></p>
  {/if}

  {#if result}
    <div class="result-card">
      <div class="result-header">
        <div class="result-label">Proposed settings.json</div>
        <button class="copy-btn" onclick={copyReview}>
          {copied ? '✓ Copied!' : 'Copy to clipboard'}
        </button>
      </div>

      <div class="review-note">
        Review this and paste it into <code>{result.path}</code> yourself,
        then verify the file is valid. Drishti never edits your file.
      </div>

      <pre class="proposed">{result.proposed}</pre>

      {#if result.added.length > 0}
        <div class="added-section">
          <div class="added-label">Keys added by this suggestion:</div>
          <div class="added-chips">
            {#each result.added as key (key)}
              <span class="added-chip">{key}</span>
            {/each}
          </div>
        </div>
      {:else}
        <p class="faint">No new keys — your settings.json already includes the statusLine entry.</p>
      {/if}

      <div class="path-row">
        <span class="path-label">Target file:</span>
        <code class="path-val">{result.path}</code>
      </div>
    </div>
  {/if}
</section>

<style>
  .panel { margin-bottom: 2rem; }
  .panel h2 { font-size: 0.95rem; color: var(--text-dim); margin: 0 0 1rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; }

  .intro { color: var(--text-faint); font-size: 0.875rem; margin: 0 0 1rem; max-width: 58ch; }
  .gen-btn {
    padding: 0.4rem 1rem; border: 1px solid var(--accent); border-radius: 6px;
    background: var(--accent-soft); color: var(--accent); font-size: 0.9rem;
    cursor: pointer; margin-bottom: 0.75rem;
    transition: background 0.15s;
  }
  .gen-btn:hover:not(:disabled) { background: var(--accent-dim); }
  .gen-btn:disabled { opacity: 0.5; cursor: not-allowed; }

  .save-err { color: var(--red); font-size: 0.85rem; margin: 0.25rem 0; }
  .retry-btn { background: none; border: 1px solid var(--red); border-radius: 5px; color: var(--red); padding: 0.15rem 0.5rem; cursor: pointer; font-size: 0.8rem; }
  .retry-btn:hover { background: var(--red-soft); }

  .result-card { background: var(--panel-2); border-radius: 10px; padding: 1rem; margin-top: 0.5rem; }
  .result-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 0.5rem; }
  .result-label { font-size: 0.82rem; color: var(--text-faint); }
  .copy-btn {
    padding: 0.28rem 0.75rem; border: 1px solid var(--border); border-radius: 6px;
    background: var(--panel); color: var(--text-dim); font-size: 0.82rem; cursor: pointer;
    transition: all 0.15s;
  }
  .copy-btn:hover { border-color: var(--accent); color: var(--accent); }

  .review-note {
    background: var(--amber-soft); border-left: 3px solid var(--amber);
    border-radius: 4px; padding: 0.45rem 0.75rem;
    color: var(--amber); font-size: 0.82rem; margin-bottom: 0.75rem;
  }

  .proposed {
    background: var(--panel); border: 1px solid var(--border-soft); border-radius: 6px;
    padding: 0.75rem; font-size: 0.82rem; overflow-x: auto;
    white-space: pre; font-family: 'IBM Plex Mono', monospace; margin: 0 0 0.75rem;
    max-height: 20rem; overflow-y: auto;
  }

  .added-section { margin-bottom: 0.75rem; }
  .added-label { font-size: 0.8rem; color: var(--text-faint); margin-bottom: 0.35rem; }
  .added-chips { display: flex; gap: 0.4rem; flex-wrap: wrap; }
  .added-chip { padding: 0.15rem 0.55rem; border: 1px solid var(--accent); border-radius: 999px; font-size: 0.8rem; color: var(--accent); background: var(--accent-soft); }

  .path-row { display: flex; align-items: center; gap: 0.5rem; }
  .path-label { font-size: 0.8rem; color: var(--text-faint); }
  .path-val { font-size: 0.82rem; color: var(--text-dim); }

  .faint { color: var(--text-faint); font-size: 0.85rem; margin: 0.25rem 0; }
</style>
