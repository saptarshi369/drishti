package services

import (
	"fmt"
	"sort"

	"github.com/saptarshi369/drishti/internal/model"
)

// Quota alert thresholds (M8 spec §7, owner-approved over the mockup's 70%):
// 80% raises an amber alert, 95% a red one.
const (
	quotaWarnPct = 80.0
	quotaCritPct = 95.0
)

// severityRank orders alerts for display: red first, then amber, then grey.
var severityRank = map[string]int{"red": 0, "amber": 1, "grey": 2}

// AlertInputs are the already-gathered signals BuildAlerts derives the current
// alert list from. Internal compute struct (never serialized). BlockedEvents are
// expected to be pre-filtered by the assembler to a recent window so stale
// blocks age out and the alert auto-clears.
type AlertInputs struct {
	Quota         model.QuotaSnapshot
	BlockedEvents []model.RecentEvent
	SecurityHigh  int
	DeadSkills    int
}

// BuildAlerts re-derives the current Overview alert list from live signals. Pure
// and stateless: alerts appear while their condition holds and vanish when it
// clears (no persistence, no acknowledge — M8 spec §7). The result is always
// non-nil so the JSON shape is stable ([] not null) for the all-clear UI.
func BuildAlerts(in AlertInputs) []model.Alert {
	alerts := []model.Alert{}

	// 1. Plan-quota threshold — only when a sample exists (gated otherwise).
	if in.Quota.Available {
		pct, ts := worstQuota(in.Quota)
		if pct >= quotaWarnPct {
			sev := "amber"
			if pct >= quotaCritPct {
				sev = "red"
			}
			alerts = append(alerts, model.Alert{
				Kind: "quota", Severity: sev,
				Text: fmt.Sprintf("Plan quota at %d%%", int(pct+0.5)),
				TsMs: ts, CTA: "usage",
			})
		}
	}

	// 2. Dangerous command blocked — one alert per recent blocked event. Names
	// only (Privacy D8): never the raw command text.
	for _, e := range in.BlockedEvents {
		text := "Command blocked"
		if e.ToolName != "" {
			text = "Command blocked — " + e.ToolName
		}
		alerts = append(alerts, model.Alert{
			Kind: "blocked_command", Severity: "red",
			Text: text, TsMs: e.TsMs, CTA: "activity",
		})
	}

	// 3. High-severity security findings present (current state, no timestamp).
	if in.SecurityHigh > 0 {
		alerts = append(alerts, model.Alert{
			Kind: "security", Severity: "red",
			Text: fmt.Sprintf("%d high-severity security finding%s", in.SecurityHigh, plural(in.SecurityHigh)),
			CTA:  "security",
		})
	}

	// 4. Dead skills present (current state, no timestamp).
	if in.DeadSkills > 0 {
		alerts = append(alerts, model.Alert{
			Kind: "dead_skill", Severity: "grey",
			Text: fmt.Sprintf("%d dead skill%s (never fired)", in.DeadSkills, plural(in.DeadSkills)),
			CTA:  "skills",
		})
	}

	// Stable order: severity rank, then newest first within a rank.
	sort.SliceStable(alerts, func(i, j int) bool {
		if severityRank[alerts[i].Severity] != severityRank[alerts[j].Severity] {
			return severityRank[alerts[i].Severity] < severityRank[alerts[j].Severity]
		}
		return alerts[i].TsMs > alerts[j].TsMs
	})
	return alerts
}

// worstQuota returns the higher used-% of the two windows and that sample's time.
func worstQuota(q model.QuotaSnapshot) (float64, int64) {
	var pct float64
	var ts int64
	for _, w := range []*model.QuotaWindow{q.FiveHour, q.SevenDay} {
		if w != nil && w.UsedPercentage > pct {
			pct = w.UsedPercentage
			ts = w.TsMs
		}
	}
	return pct, ts
}

// plural returns "s" for counts other than 1 (English pluralisation helper).
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
