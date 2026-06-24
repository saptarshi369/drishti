package services

import (
	"fmt"
	"time"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/skills"
	"github.com/saptarshi369/drishti/internal/store"
)

// Overview is the assembled payload for the Overview screen / SSE "counters"
// frame. KPIs + Recent are global (per agent); the rest are the M8 aggregates
// scoped to one root. It is the single source the API and the SSE broadcaster
// both use, so the page and the live stream never disagree.
type Overview struct {
	KPIs             model.OverviewKPIs     `json:"kpis"`
	Recent           []model.RecentEvent    `json:"recent"`
	ActiveComponents model.ActiveComponents `json:"active_components"`
	ActiveHere       []model.ActiveHereRow  `json:"active_here"`
	ContextTax       model.ContextTax       `json:"context_tax"`
	Health           model.HealthSnapshot   `json:"health"`
	Alerts           []model.Alert          `json:"alerts"`
}

// OverviewParams carries the per-server settings the assembler needs but cannot
// read itself (they live on the API server behind its config mutex): the watched
// root, the context-window denominator, and the skill-analytics thresholds. The
// caller fills these via the M7 accessors (never the raw fields).
type OverviewParams struct {
	Root            string
	WindowTokens    int
	SkillThresholds skills.Thresholds
}

// blockedAlertWindowMs bounds how recent a blocked event must be to raise an
// alert, so stale blocks age out and the alert auto-clears (M8 spec §7).
const blockedAlertWindowMs = 60 * 60 * 1000 // 60 minutes

// OverviewSnapshot assembles the full Overview payload. The baseline (KPIs,
// recent events) returns an error on failure — those are the slice guarantees
// the page cannot do without. It is READ-ONLY: cost is stamped at ingest, so the
// 1s broadcast never takes the write lock. Every per-root aggregate degrades
// independently (§14): a failing section leaves its zero/empty value and the rest
// of the snapshot is still returned, mirroring snapshotMessages' activity/quota
// handling.
func OverviewSnapshot(st *store.Store, p OverviewParams) (Overview, error) {
	// NOTE: no cost backfill here. est_cost_usd is stamped at ingest (store.SetCostFn),
	// so this hot read/broadcast path is read-only and never contends the write lock.
	kpis, err := st.OverviewKPIs()
	if err != nil {
		return Overview{}, err
	}
	recent, err := st.RecentEvents(20)
	if err != nil {
		return Overview{}, err
	}
	ov := Overview{KPIs: kpis, Recent: recent, Alerts: []model.Alert{}}
	// Initialize ActiveHere to a non-nil empty slice so it serializes as []
	// not null when ListResolved errors or returns no rows — matching the
	// non-nil guarantee already applied to Alerts above (JSON-contract D8).
	ov.ActiveHere = []model.ActiveHereRow{}

	// Resolved inventory → active-components, "Active here", context-tax. One read
	// (showDisabled=true) serves all three; the helpers keep only active rows.
	var ctxPct float64
	var hooksConfigured bool
	if rows, lerr := st.ListResolved("", p.Root, true); lerr == nil {
		ov.ActiveComponents, ov.ActiveHere = buildActiveInventory(rows)
		cb := BuildContextBudget(rows, p.WindowTokens)
		ov.ContextTax = model.ContextTax{TotalTokens: cb.TotalTokens, WindowTokens: cb.WindowTokens, Pct: cb.Pct}
		ctxPct = cb.Pct
		for _, r := range rows {
			if r.EffectiveStatus == "active" && r.Category == "hook" {
				hooksConfigured = true
				break
			}
		}
		// Append the context-tax row to "Active here".
		ov.ActiveHere = append(ov.ActiveHere, model.ActiveHereRow{
			Category: "context", Count: cb.TotalTokens, Note: contextNote(cb.Pct), CTA: "context",
		})
	}

	// Security findings → health + alert input.
	secCounts := map[string]int{}
	if findings, ferr := st.ListFindings("claude", p.Root); ferr == nil {
		secCounts = BuildSecuritySnapshot(findings).Counts
	}

	// Skill analytics → health + alert input.
	var skillCounts model.SkillCounts
	var activeSkills int
	if srows, serr := st.SkillAnalytics(p.Root); serr == nil {
		ss := skills.BuildAnalytics(srows, p.SkillThresholds)
		skillCounts = ss.Counts
		activeSkills = ss.Total - ss.Counts.Disabled
	}

	// Recent execution errors (24h) for hook-health (blocks are NOT errors).
	now := time.Now().UnixMilli()
	dayAgo := now - 24*int64(time.Hour/time.Millisecond)
	var hookErrors int
	if c, cerr := st.ActivityCounters(dayAgo, ""); cerr == nil {
		hookErrors = c.Errors
	}

	ov.Health = BuildHealth(HealthInputs{
		ContextPct:      ctxPct,
		SecurityCounts:  secCounts,
		SkillCounts:     skillCounts,
		ActiveSkills:    activeSkills,
		HookErrors:      hookErrors,
		HooksConfigured: hooksConfigured,
	})

	// Alerts: quota + recent blocked + high security + dead skills.
	quota, _ := QuotaSnapshot(st, "claude") // err → zero value (Available=false) → no quota alert
	var blocked []model.RecentEvent
	if evs, eerr := st.EventsPage("blocked", 5, 0); eerr == nil {
		for _, e := range evs {
			if now-e.TsMs <= blockedAlertWindowMs {
				blocked = append(blocked, e)
			}
		}
	}
	ov.Alerts = BuildAlerts(AlertInputs{
		Quota:         quota,
		BlockedEvents: blocked,
		SecurityHigh:  secCounts["high"],
		DeadSkills:    skillCounts.Dead,
	})
	return ov, nil
}

// contextNote renders the "N% of window" note for the Active-here context row.
func contextNote(pct float64) string {
	return fmt.Sprintf("%d%% of window", int(pct+0.5))
}
