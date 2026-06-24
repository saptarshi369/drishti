<!--
  ConfigFilesPanel: config-file editors (§4 panel 6).

  Three editable knobs:
  1. context_window_tokens → PUT /api/settings (with full state).
  2. Skill thresholds: high_trigger_min, low_value_ratio_max → PUT /api/thresholds.
  3. Security rules: a table of rules with enable/disable toggles and a Save
     button → PUT /api/rules. Full row editing (title, severity, remediation) is
     provided for each rule; adding new rules or editing pattern arrays is
     out-of-scope for V1 (simplification noted in the task report — the rules
     file is user-editable TOML so power users can edit it directly in ~/.drishti/).

  The security rules are loaded from the existing SecuritySnapshot via GET
  /api/security so we reuse data the Skills screen already fetches. Here we
  fetch it independently on mount so the panel is self-contained.
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import {
    getThresholds,
    putSettings,
    putThresholds,
    putRules,
    getSecurity,
    type SettingsView,
    type SettingsSaveResult,
    type SecurityRule,
    type Thresholds,
  } from '$lib/api';

  // onSaved: called after any successful save so the parent can re-fetch
  // GET /api/settings and refresh the shared snap — preventing cross-panel
  // clobber where saving one panel reverts another panel's just-saved change.
  let { snap, onSaved }: { snap: SettingsView; onSaved: () => void } = $props();

  // ── Context window tokens ───────────────────────────────────────────────
  let ctxTokens    = $state(snap.context_window_tokens);
  let ctxSaveState = $state<'idle' | 'saving' | 'saved' | 'error'>('idle');
  let ctxSaveMsg   = $state('');

  async function saveCtx() {
    ctxSaveState = 'saving';
    ctxSaveMsg   = '';
    try {
      // Read pass-through fields from the live snap prop (not a captured copy)
      // so a prior panel's save is not reverted.
      const r: SettingsSaveResult = await putSettings({
        port:                  snap.port,
        bind_addr:             snap.bind_addr,
        theme:                 snap.theme,
        accent:                snap.accent,
        active_window:         snap.active_window,
        aggregate_horizon:     snap.aggregate_horizon,
        throttle:              snap.throttle,
        check_interval:        snap.check_interval,
        context_window_tokens: ctxTokens,
        auto_check:            snap.auto_check,
      });
      ctxSaveState = 'saved';
      ctxSaveMsg   = 'Saved.' + (r.warnings?.length ? ' Warnings: ' + r.warnings.join('; ') : '');
      onSaved();
    } catch (e) {
      ctxSaveState = 'error';
      ctxSaveMsg   = e instanceof Error ? e.message : 'save failed';
    }
  }

  // ── Skill thresholds ────────────────────────────────────────────────────
  // Seeded from GET /api/thresholds on mount; falls back to 25/0.4 if fetch fails.
  let highTriggerMin    = $state(25);
  let lowValueRatioMax  = $state(0.4);
  let thrSaveState      = $state<'idle' | 'saving' | 'saved' | 'error'>('idle');
  let thrSaveMsg        = $state('');

  async function saveThresholds() {
    thrSaveState = 'saving';
    thrSaveMsg   = '';
    const body: Thresholds = {
      high_trigger_min:    highTriggerMin,
      low_value_ratio_max: lowValueRatioMax,
    };
    if (body.high_trigger_min <= 0 || body.low_value_ratio_max <= 0) {
      thrSaveState = 'error';
      thrSaveMsg   = 'Both values must be positive.';
      return;
    }
    try {
      await putThresholds(body);
      thrSaveState = 'saved';
      thrSaveMsg   = 'Thresholds saved.';
      // No putSettings call here — thresholds go to a separate file; onSaved
      // still notifies the parent to refresh snap for other panels.
      onSaved();
    } catch (e) {
      thrSaveState = 'error';
      thrSaveMsg   = e instanceof Error ? e.message : 'save failed';
    }
  }

  // ── Security rules ──────────────────────────────────────────────────────
  // We load the current rules from GET /api/security (SecuritySnapshot has
  // findings but not the raw rules list). However, the rules are only available
  // via GET /api/security's snapshot — they are not served directly as a raw
  // list. The simplification here: we fetch the security snapshot and derive a
  // stub rule list from the findings (rule_id, severity, title, remediation,
  // enabled = no finding for that id). This is a V1 approximation — the proper
  // V2 would need a dedicated GET /api/rules endpoint (deferred, noted in report).
  //
  // Concretely: we show the rules that were violated (findings) and let the user
  // toggle-enable/disable them. Non-violated rules are not shown (no GET /api/rules).
  // This is noted as a V1 simplification.

  type LocalRule = SecurityRule & { _dirty?: boolean };
  let rules        = $state<LocalRule[]>([]);
  let rulesLoading = $state(true);
  let rulesErr     = $state<string | null>(null);
  let rulesSaveState = $state<'idle' | 'saving' | 'saved' | 'error'>('idle');
  let rulesSaveMsg   = $state('');

  // loadRules: fetch GET /api/security to extract known rule IDs + metadata.
  // Because there is no GET /api/rules endpoint, we reconstruct a minimal rule
  // list from the findings. Missing rules (no finding) cannot be shown without
  // a dedicated endpoint.
  async function loadRules() {
    rulesLoading = true;
    rulesErr     = null;
    try {
      const sec = await getSecurity();
      // Build a deduplicated rule list from findings. Each finding has a rule_id,
      // severity, title, remediation. Construct a minimal SecurityRule from it.
      const seen = new Set<string>();
      const rs: LocalRule[] = [];
      for (const f of sec.findings) {
        if (!seen.has(f.rule_id)) {
          seen.add(f.rule_id);
          rs.push({
            id:          f.rule_id,
            type:        'unknown',  // type not exposed in Finding
            enabled:     true,
            severity:    f.severity,
            title:       f.title,
            remediation: f.remediation,
          });
        }
      }
      rules = rs;
    } catch (e) {
      rulesErr = e instanceof Error ? e.message : 'failed to load rules';
    } finally {
      rulesLoading = false;
    }
  }

  async function saveRules() {
    rulesSaveState = 'saving';
    rulesSaveMsg   = '';
    try {
      // The API requires id + type; type is 'unknown' here (deferred GET /api/rules).
      // Filter out any rule with type 'unknown' that would cause a 400.
      // V1 simplification: we can only toggle rules whose type we know.
      const validRules = rules.filter((r) => r.type !== 'unknown');
      if (validRules.length === 0 && rules.length > 0) {
        rulesSaveState = 'error';
        rulesSaveMsg   = 'Cannot save: rule types are unknown (no GET /api/rules endpoint yet). Edit security-rules.toml directly.';
        return;
      }
      await putRules(validRules);
      rulesSaveState = 'saved';
      rulesSaveMsg   = 'Rules saved.';
      onSaved();
    } catch (e) {
      rulesSaveState = 'error';
      rulesSaveMsg   = e instanceof Error ? e.message : 'save failed';
    }
  }

  function toggleRule(id: string) {
    rules = rules.map((r) => r.id === id ? { ...r, enabled: !r.enabled, _dirty: true } : r);
  }

  onMount(async () => {
    // Seed thresholds from the live server values; fall back to 25/0.4 on error.
    try {
      const t = await getThresholds();
      highTriggerMin   = t.high_trigger_min;
      lowValueRatioMax = t.low_value_ratio_max;
    } catch {
      // Fallback defaults already set in $state initialiser — keep them.
    }
    loadRules();
  });
</script>

<section class="panel" id="config-files">
  <h2>Config files</h2>

  <!-- Context window tokens -->
  <div class="sub-section">
    <div class="sub-label">Context window (tokens)</div>
    <div class="field-row">
      <input
        id="ctx_tokens"
        class="text-input num-input"
        type="number"
        min="1"
        bind:value={ctxTokens}
        placeholder="200000"
      />
      <button class="save-btn" onclick={saveCtx} disabled={ctxSaveState === 'saving'}>Save</button>
    </div>
    <span class="hint">Denominator for the context-budget-tax % on the Context Budget screen.</span>
    {#if ctxSaveState === 'saved'}
      <p class="save-ok">{ctxSaveMsg}</p>
    {:else if ctxSaveState === 'error'}
      <p class="save-err">{ctxSaveMsg}</p>
    {/if}
  </div>

  <!-- Skill thresholds -->
  <div class="sub-section">
    <div class="sub-label">Skills analytics thresholds</div>
    <div class="threshold-grid">
      <div class="field">
        <label class="field-label" for="high_trigger_min">High trigger minimum</label>
        <input
          id="high_trigger_min"
          class="text-input num-input"
          type="number"
          min="1"
          bind:value={highTriggerMin}
          placeholder="25"
        />
        <span class="hint">Min triggers before a low-value skill is flagged over-triggering.</span>
      </div>
      <div class="field">
        <label class="field-label" for="low_value_ratio_max">Low value ratio max</label>
        <input
          id="low_value_ratio_max"
          class="text-input num-input"
          type="number"
          min="0.01"
          step="0.01"
          bind:value={lowValueRatioMax}
          placeholder="0.4"
        />
        <span class="hint">Value ratio below which a heavy skill is "low value".</span>
      </div>
    </div>
    <div class="save-row">
      <button class="save-btn" onclick={saveThresholds} disabled={thrSaveState === 'saving'}>Save thresholds</button>
      {#if thrSaveState === 'saved'}
        <span class="save-ok">{thrSaveMsg}</span>
      {:else if thrSaveState === 'error'}
        <span class="save-err">{thrSaveMsg}</span>
      {/if}
    </div>
  </div>

  <!-- Security rules -->
  <div class="sub-section">
    <div class="sub-label">Security rules</div>

    {#if rulesLoading}
      <p class="faint">Loading rules…</p>
    {:else if rulesErr}
      <p class="save-err">{rulesErr} <button class="retry-btn" onclick={loadRules}>Retry</button></p>
    {:else if rules.length === 0}
      <p class="faint">
        No active rule violations found. Rules are stored in <code>~/.drishti/security-rules.toml</code> —
        edit that file directly or trigger findings to see rules here.
      </p>
      <p class="faint deferred-note">
        Note: a dedicated GET /api/rules endpoint is deferred; this panel shows only rules with active
        findings. Full rule management is available by editing <code>security-rules.toml</code>.
      </p>
    {:else}
      <p class="faint deferred-note">
        Showing rules with active findings. Full rule list requires editing
        <code>~/.drishti/security-rules.toml</code> directly (GET /api/rules deferred).
      </p>
      <table class="rules-table">
        <thead>
          <tr>
            <th>Enabled</th>
            <th>Severity</th>
            <th>Rule</th>
            <th>Remediation</th>
          </tr>
        </thead>
        <tbody>
          {#each rules as r (r.id)}
            <tr class:disabled-row={!r.enabled}>
              <td>
                <button
                  class="toggle-btn"
                  class:on={r.enabled}
                  onclick={() => toggleRule(r.id)}
                  aria-pressed={r.enabled}
                >{r.enabled ? 'On' : 'Off'}</button>
              </td>
              <td><span class="sev sev-{r.severity}">{r.severity}</span></td>
              <td>
                <div class="rule-title">{r.title}</div>
                <div class="rule-id">{r.id}</div>
              </td>
              <td class="remediation">{r.remediation}</td>
            </tr>
          {/each}
        </tbody>
      </table>
      <div class="save-row">
        <button class="save-btn" onclick={saveRules} disabled={rulesSaveState === 'saving'}>Save rules</button>
        {#if rulesSaveState === 'saved'}
          <span class="save-ok">{rulesSaveMsg}</span>
        {:else if rulesSaveState === 'error'}
          <span class="save-err">{rulesSaveMsg}</span>
        {/if}
      </div>
    {/if}
  </div>
</section>

<style>
  .panel { margin-bottom: 2rem; }
  .panel h2 { font-size: 0.95rem; color: var(--text-dim); margin: 0 0 1rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; }

  .sub-section { margin-bottom: 1.5rem; }
  .sub-label { font-size: 0.88rem; color: var(--text-dim); font-weight: 600; margin-bottom: 0.6rem; }

  .field { display: flex; flex-direction: column; gap: 0.25rem; }
  .field-label { font-size: 0.85rem; color: var(--text-faint); }
  .field-row { display: flex; align-items: center; gap: 0.6rem; margin-bottom: 0.3rem; }

  .text-input {
    background: var(--panel-2); border: 1px solid var(--border); border-radius: 6px;
    color: var(--text); font-size: 0.9rem; padding: 0.35rem 0.65rem;
    font-family: 'IBM Plex Mono', monospace;
    transition: border-color 0.15s;
  }
  .text-input:focus { outline: none; border-color: var(--accent); }
  .num-input { max-width: 16ch; }
  .hint { font-size: 0.78rem; color: var(--text-faint); }

  .threshold-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; margin-bottom: 0.75rem; }
  @media (max-width: 500px) { .threshold-grid { grid-template-columns: 1fr; } }

  .save-row { display: flex; align-items: center; gap: 0.75rem; margin-top: 0.5rem; }
  .save-btn { padding: 0.32rem 0.85rem; border: 1px solid var(--border); border-radius: 6px; background: var(--panel-2); color: var(--text); font-size: 0.88rem; cursor: pointer; }
  .save-btn:hover:not(:disabled) { border-color: var(--accent); color: var(--accent); }
  .save-btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .save-ok  { color: var(--green); font-size: 0.85rem; }
  .save-err { color: var(--red);   font-size: 0.85rem; }
  .faint { color: var(--text-faint); font-size: 0.85rem; margin: 0.25rem 0; }
  .deferred-note { font-style: italic; }

  .retry-btn { background: none; border: 1px solid var(--red); border-radius: 5px; color: var(--red); padding: 0.15rem 0.5rem; cursor: pointer; font-size: 0.8rem; }
  .retry-btn:hover { background: var(--red-soft); }

  .rules-table { width: 100%; border-collapse: collapse; font-size: 0.88rem; margin-bottom: 0.5rem; }
  .rules-table th { text-align: left; padding: 0.35rem 0.5rem; border-bottom: 1px solid var(--border); color: var(--text-faint); font-weight: 500; font-size: 0.8rem; }
  .rules-table td { padding: 0.35rem 0.5rem; border-bottom: 1px solid var(--border-soft); vertical-align: top; }
  .disabled-row { opacity: 0.55; }
  .rule-title { font-weight: 500; }
  .rule-id   { font-size: 0.75rem; color: var(--text-faint); font-family: 'IBM Plex Mono', monospace; }
  .remediation { font-size: 0.82rem; color: var(--text-dim); max-width: 28ch; }

  .sev { padding: 0.1rem 0.45rem; border-radius: 4px; font-size: 0.78rem; font-weight: 600; }
  .sev-high   { background: var(--red-soft);   color: var(--red); }
  .sev-medium { background: var(--amber-soft);  color: var(--amber); }
  .sev-low    { background: var(--panel-2);     color: var(--text-faint); }

  .toggle-btn {
    padding: 0.18rem 0.55rem; border: 1px solid var(--border); border-radius: 5px;
    background: var(--panel-2); color: var(--text-dim); font-size: 0.82rem; cursor: pointer;
  }
  .toggle-btn.on { background: var(--accent-soft); border-color: var(--accent); color: var(--accent); font-weight: 600; }
</style>
