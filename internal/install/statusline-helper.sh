#!/usr/bin/env bash
# Drishti statusline helper (manual install).
#
# Claude Code invokes the configured `statusLine` command on every status
# refresh, piping a JSON blob on stdin that includes the plan rate_limits. This
# script forwards those numbers to the local Drishti daemon and prints a minimal
# status line so your prompt still renders.
#
# Install (manual; the consent/backup auto-installer is Drishti's Settings module):
#   1. Make executable:   chmod +x scripts/statusline-helper.sh
#   2. In ~/.claude/settings.json add:
#        "statusLine": { "type": "command",
#                        "command": "/ABSOLUTE/PATH/TO/scripts/statusline-helper.sh" }
#
# Notes:
#   - Fails silently if the daemon is down (never breaks your status line).
#   - Requires `jq` and `curl` (both common). Without jq it still prints a line.
#   - Posts to 127.0.0.1:7777 (Drishti's default port). Edit DRISHTI_URL if you
#     run the daemon on a different port.
#
# Claude Code statusLine stdin JSON shape (confirmed via Claude Code docs):
#   .rate_limits.five_hour.used_percentage  — float, 0-100
#   .rate_limits.five_hour.resets_at        — Unix epoch SECONDS (not ms)
#   .rate_limits.seven_day.used_percentage  — float, 0-100
#   .rate_limits.seven_day.resets_at        — Unix epoch SECONDS (not ms)
#   .plan                                    — string, e.g. "max"
#
# NOTE: The brief assumed `resets_at_ms` but the real field is `resets_at` in
# seconds. This script reads `resets_at` and multiplies by 1000 before posting
# to the daemon's POST /api/quota/sample, which expects `resets_at_ms`.

set -u
DRISHTI_URL="${DRISHTI_URL:-http://127.0.0.1:7777}"

# Read the whole stdin payload once.
payload="$(cat)"

if command -v jq >/dev/null 2>&1; then
  # Extract the two windows. `// empty` yields nothing when a path is absent,
  # which we translate into an omitted JSON key below.
  five_pct="$(printf '%s' "$payload"  | jq -r '.rate_limits.five_hour.used_percentage // empty' 2>/dev/null)"
  # resets_at is epoch seconds; multiply by 1000 for the daemon's resets_at_ms.
  five_rst_s="$(printf '%s' "$payload" | jq -r '.rate_limits.five_hour.resets_at // empty' 2>/dev/null)"
  week_pct="$(printf '%s' "$payload"  | jq -r '.rate_limits.seven_day.used_percentage // empty' 2>/dev/null)"
  week_rst_s="$(printf '%s' "$payload" | jq -r '.rate_limits.seven_day.resets_at // empty' 2>/dev/null)"
  plan="$(printf '%s' "$payload"      | jq -r '.plan // empty' 2>/dev/null)"

  # Convert resets_at (seconds) -> resets_at_ms (milliseconds) for the daemon.
  # Fall back to 0 if the field was absent.
  five_rst_ms=$(( ${five_rst_s:-0} * 1000 ))
  week_rst_ms=$(( ${week_rst_s:-0} * 1000 ))

  # Build the body only from the windows that are present.
  body='{"agent":"claude","source":"statusline"'
  [ -n "$plan" ] && body="$body,\"plan\":\"$plan\""
  if [ -n "$five_pct" ]; then
    body="$body,\"five_hour\":{\"used_percentage\":$five_pct,\"resets_at_ms\":$five_rst_ms}"
  fi
  if [ -n "$week_pct" ]; then
    body="$body,\"seven_day\":{\"used_percentage\":$week_pct,\"resets_at_ms\":$week_rst_ms}"
  fi
  body="$body}"

  # Forward, best-effort. -m 1 keeps the status line snappy; errors are swallowed.
  curl -s -m 1 -X POST -H 'Content-Type: application/json' \
    -d "$body" "$DRISHTI_URL/api/quota/sample" >/dev/null 2>&1 || true

  # Print a compact status line.
  printf 'drishti · 5h %s%% · 7d %s%%' "${five_pct:-–}" "${week_pct:-–}"
else
  # No jq: don't forward, just print a minimal line so the status bar isn't empty.
  printf 'drishti'
fi
