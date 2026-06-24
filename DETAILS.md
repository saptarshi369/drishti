# Drishti — screen-by-screen guide

This guide explains what each screen shows, how to read it, and how to tune it. For installation and
configuration basics, see the [README](README.md).

Drishti runs at `http://127.0.0.1:7777`. Every screen updates live — you rarely need to refresh.
A status dot in the app shell tells you whether the daemon is live, starting, or offline.

---

## Overview (Command Center)

Your at-a-glance home screen. It aggregates every other screen into one view.

- **KPI cards** — *Active components* (how many skills, MCP servers, hooks, and agents are active
  right now), *Prompts today*, *Spend today* (estimated), and *Plan quota*. Each card links to the
  screen with the full story.
- **Live activity** — a sparkline of recent prompt rate plus the latest events as they happen.
- **Harness health** — a single 0–100 score made of four sub-scores:
  - *Context tax* — lower is better; high means your always-on components eat a lot of the window.
  - *Security* — drops as the audit finds more (and more severe) issues.
  - *Skill hygiene* — drops when you have dead or over-triggering skills.
  - *Hook health* — drops when hooks are configured and recent runs produced errors. (This is an
    approximate signal; a hook *blocking* a bad command is not penalized.)
- **Active here** — a quick resolved summary of what's active in the watched project, each row
  linking into the Inventory or Context screens.
- **Alerts** — surfaced when something needs attention: plan quota crossing 80% / 95%, a command
  blocked, a high-severity security finding, or dead skills present. Alerts clear themselves when
  the condition goes away.

---

## Inventory (Harness Map)

The source of truth for *what is active*. Claude Code merges configuration across user and project
scope; Drishti resolves the winner for every component and shows you why.

- **Category tabs** — skills, MCP servers, hooks, agents, commands, output-styles, memory, and
  plugins.
- **Scope resolution** — each item shows where it came from (user vs project) and its effective
  status.
- **"Why?" trail** — open an item to see its full precedence trail: which definitions existed,
  which won, and which were overridden or shadowed.
- **Show disabled** — toggle to include items that are present but not in effect.

Use this screen whenever you're surprised that something *is* or *isn't* active — the trail tells
you exactly which file decided it.

---

## Live Activity

A real-time stream of what Claude Code is doing.

- **Counters** — prompts, skills, tool calls, blocked actions, and errors, shown for both the
  current session and the rolling 24 hours.
- **Sparklines** — per-minute rates so you can see bursts.
- **Event stream** — the most recent events as they arrive, newest first.
- **Skills triggered** — which skills have fired, with a warning marker on any active skill that
  has *never* fired (a candidate to archive).

Everything here streams over Server-Sent Events; leave it open and watch it move.

---

## Usage & Cost

Where your tokens and dollars go. Costs are **estimates** from a local pricing table.

- **Trend** — the last 7 days of input / output / cache tokens and estimated USD. Toggle between a
  **Bars** and a **Chart** view (your choice persists).
- **By project / by model** — which repositories and which models (Opus / Sonnet / Haiku) drive
  your spend.
- **Heatmap + streak** — an activity heatmap over recent weeks and your current daily streak.
- **Plan quota** — session and weekly gauges. These are **gated** until you install the statusline
  helper (below); until then they show an install prompt.

### Enabling plan-quota gauges

Claude Code can report your plan rate-limits through its status line. Drishti ships a small helper
that forwards those numbers:

1. Open **Settings → Live helper** and click *Generate suggestion*. Drishti produces a ready-to-paste
   `statusLine` snippet for your `settings.json` — it **never writes the file for you**.
2. Paste it into your Claude Code `settings.json` yourself.

The gauges go live on the next status-line update. (Or wire `scripts/statusline-helper.sh` manually.)

---

## Context Budget

The "context tax" — how much of your context window is consumed by always-on components *before you
type anything*.

- **Total tax + % of window** — the headline number and its share of the window.
- **By category** — a stacked breakdown (skills, MCP, hooks, memory, …).
- **Biggest consumers** — the individual items costing the most.
- **"If disabled" recompute** — uncheck a row to see what your tax *would* be without it. This is a
  live, client-side estimate to help you decide what to trim.

> Note: MCP token costs are an approximation (a flat per-server estimate); a caveat is shown when an
> MCP server is configured.

**Tune it:** the window size used as the percentage denominator is `[context] window_tokens` in
`~/.drishti/config.toml` (default 200000).

---

## Security & Audit

A configurable audit of your Claude Code configuration. Findings are grouped by severity
(**high / medium / low**), each with a title, the target it refers to, a detail, and a suggested
remediation. A clean config shows an all-clear state.

The default rules flag things like: missing deny rules for sensitive paths, bypass permission
modes, overly broad allow rules, secrets in MCP environment variables, and untrusted plugin
sources.

**Privacy:** Drishti detects secret-shaped values only to *count* them — it never reads, stores, or
displays the values themselves. Only key names appear.

**Tune it:** the rules live in `~/.drishti/security-rules.toml`, a plain, fully commented file. You
can disable a rule, change its parameters, or add new entries (for example, a new path that must
have a deny rule). Save the file and Drishti applies your changes within ~10 seconds — no restart.

---

## Skills Analytics

Whether your skills earn their keep. Skills are always-on context cost, so this screen helps you
prune.

- **Sortable table** — by triggers, context cost, value ratio, or name.
- **Value ratio** — triggers per 1,000 tokens of always-on cost (higher is better; shown as "—" for
  zero-cost rows).
- **Badges** — *dead* (active but never fired), *over-triggering* (fires a lot for little value),
  and *disabled*.
- **Summary chips** — total skills, total context tokens, and the counts of each flag.

**Tune it:** the thresholds that decide *dead* and *over-triggering* live in
`~/.drishti/skills-analytics.toml` (e.g. `high_trigger_min`, `low_value_ratio_max`). Edit and save;
changes apply within ~10 seconds.

---

## Settings

Tune Drishti itself, from the browser. All changes are saved to `~/.drishti/` only.

- **Appearance** — theme and accent, applied live.
- **Watched folders** — a folder picker (constrained to your home directory) to choose which roots
  Drishti scans.
- **Retention & daemon** — how long to keep data, and the port / bind address. Changing the port or
  bind address shows a *restart required* banner.
- **Updates** — an opt-in "Check for updates" (the only outbound network call).
- **Config files** — edit the skills-analytics thresholds inline.
- **Live helper** — generate the statusline snippet for plan-quota (copy-to-clipboard; never writes
  your `settings.json`).
- **MCP servers** — a read-only list of the MCP servers Drishti found.

---

## The tuning files

Everything Drishti owns lives in `~/.drishti/`:

| File | What it controls |
|------|------------------|
| `config.toml` | Port, bind address, watched roots, retention, context-window size, appearance, update check. |
| `security-rules.toml` | The Security & Audit rules. |
| `skills-analytics.toml` | The dead / over-triggering thresholds. |

The two rule files are reloaded automatically within ~10 seconds of a save — no restart. Most of
`config.toml` is also hot-applied; only the port and bind address require a restart.

**Reset:** Drishti never touches your Claude Code config. To reset Drishti itself, stop the daemon
and delete `~/.drishti/` — it will be recreated on next launch.
