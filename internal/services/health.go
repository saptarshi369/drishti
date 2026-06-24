package services

import "github.com/saptarshi369/drishti/internal/model"

// Health sub-score penalty weights — the only tunable constants in the composite.
// Kept here (documented) so a future change is a one-line edit, not a hunt
// through the formula. See the M8 spec §6.
const (
	secHighPenalty = 25 // points removed per high-severity security finding
	secMedPenalty  = 10 // …per medium finding
	secLowPenalty  = 3  // …per low finding
	hookErrPenalty = 15 // …per recent execution error (only when hooks configured)
)

// HealthInputs are the already-gathered signals BuildHealth folds into the
// composite. It is an internal compute struct (never serialized): the assembler
// collects each signal from its own store read and hands them over in one value.
type HealthInputs struct {
	ContextPct      float64        // % of the context window consumed by always-on items
	SecurityCounts  map[string]int // findings per severity: "high","medium","low"
	SkillCounts     model.SkillCounts
	ActiveSkills    int  // active skill total (the skill-hygiene denominator)
	HookErrors      int  // recent execution-error events (NOT intentional blocks)
	HooksConfigured bool // whether any hook is active in the resolved inventory
}

// BuildHealth folds the four sub-scores into the 0–100 harness-health composite.
// Each sub-score is an integer 0–100; the ring Score is their equal-weighted
// average rounded to nearest (the simplest honest default — M8 spec §6). Pure:
// inputs in, snapshot out, no I/O.
func BuildHealth(in HealthInputs) model.HealthSnapshot {
	// Context-tax: lower always-on % → healthier. +0.5 rounds the float to nearest.
	ctx := clampScore(100 - int(in.ContextPct+0.5))
	// Security: start at 100, subtract a weighted penalty per finding, floor at 0.
	sec := clampScore(100 - (in.SecurityCounts["high"]*secHighPenalty +
		in.SecurityCounts["medium"]*secMedPenalty +
		in.SecurityCounts["low"]*secLowPenalty))
	hyg := skillHygiene(in.ActiveSkills, in.SkillCounts)
	hook := hookHealth(in.HooksConfigured, in.HookErrors)

	bars := []model.HealthBar{
		{Label: "Context tax", Score: ctx},
		{Label: "Security", Score: sec},
		{Label: "Skill hygiene", Score: hyg},
		{Label: "Hook health", Score: hook},
	}
	// +2 before the /4 integer division rounds the average to the nearest whole.
	return model.HealthSnapshot{Score: (ctx + sec + hyg + hook + 2) / 4, Bars: bars}
}

// skillHygiene scores the share of active skills that are neither dead nor
// over-triggering. With no active skills there is nothing to keep clean → 100.
func skillHygiene(active int, c model.SkillCounts) int {
	if active <= 0 {
		return 100
	}
	healthy := active - c.Dead - c.OverTriggering
	if healthy < 0 {
		healthy = 0
	}
	// +active/2 rounds (healthy/active)*100 to the nearest whole percent.
	return clampScore((healthy*100 + active/2) / active)
}

// hookHealth approximates harness "hook health" as the absence of recent
// execution errors while hooks are configured. A hook correctly BLOCKING a bad
// command is the hook working (blocks are excluded upstream, never penalised); a
// hook/tool ERRORING is the proxy for trouble. With no hooks configured nothing
// can misfire → 100. This is a documented proxy (no hook timing data exists) —
// see M8 spec §6.
func hookHealth(configured bool, errors int) int {
	if !configured {
		return 100
	}
	return clampScore(100 - errors*hookErrPenalty)
}

// clampScore bounds n to the 0–100 sub-score range.
func clampScore(n int) int {
	if n < 0 {
		return 0
	}
	if n > 100 {
		return 100
	}
	return n
}
