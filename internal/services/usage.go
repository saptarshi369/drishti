package services

import (
	"sort"
	"strings"
	"time"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/store"
)

// usageWindowDays is the trend + breakdown window; heatWindowDays is the longer
// fixed window for the activity heatmap/streak (both read the persistent rollup).
const (
	usageWindowDays = 7
	heatWindowDays  = 56
	maxProjects     = 8
)

// todayDay returns today's local date as yyyymmdd.
func todayDay() int {
	n := time.Now()
	return n.Year()*10000 + int(n.Month())*100 + n.Day()
}

// dayToTime converts a yyyymmdd int into a local midnight time.Time. It is the
// inverse used to step day-by-day and to format weekday labels.
func dayToTime(day int) time.Time {
	y := day / 10000
	m := (day / 100) % 100
	d := day % 100
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local)
}

// dayInt converts a time.Time to a yyyymmdd int (local). The inverse of
// dayToTime; used when stepping forward/backward to look up daily data.
func dayInt(t time.Time) int {
	return t.Year()*10000 + int(t.Month())*100 + t.Day()
}

// heatBucket maps a day's token total to an intensity 0-3 relative to the busiest
// day (max). 0 means no activity; max==0 always yields 0 (never divides by zero).
func heatBucket(total, max int64) int {
	if total <= 0 || max <= 0 {
		return 0
	}
	ratio := float64(total) / float64(max)
	switch {
	case ratio <= 0.33:
		return 1
	case ratio <= 0.66:
		return 2
	default:
		return 3
	}
}

// computeStreak counts consecutive active days ending at today, or at yesterday
// when today itself has no activity yet. active maps yyyymmdd → true for any day
// with total_tokens > 0.
func computeStreak(today int, active map[int]bool) int {
	cur := dayToTime(today)
	// If today has no activity, start counting from yesterday so an in-progress
	// day that hasn't logged tokens yet doesn't reset a real streak.
	if !active[dayInt(cur)] {
		cur = cur.AddDate(0, 0, -1)
	}
	streak := 0
	for active[dayInt(cur)] {
		streak++
		cur = cur.AddDate(0, 0, -1)
	}
	return streak
}

// projectLabel turns an encoded project key (e.g. "-Users-me-dev-myapp") into a
// human label — the last '-' segment ("myapp"). Empty key → "(all projects)".
func projectLabel(root string) string {
	if root == "" {
		return "(all projects)"
	}
	parts := strings.Split(root, "-")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return root
}

// modelLabel turns a model id into a short display name. Unknown ids pass through.
func modelLabel(m string) string {
	switch {
	case m == "":
		return "(unknown)"
	case strings.Contains(m, "opus"):
		return "Opus"
	case strings.Contains(m, "sonnet"):
		return "Sonnet"
	case strings.Contains(m, "haiku"):
		return "Haiku"
	default:
		return m
	}
}

// UsageSnapshot assembles the full Usage & Cost payload: a 7-day token/cost trend
// (zero-filled), window totals, by-project and by-model breakdowns, and a 56-day
// activity heatmap + current streak. Cost is back-filled first (idempotent) so the
// numbers are correct even without the scheduler. A store error short-circuits to
// a zero snapshot (failsafe, §14).
func UsageSnapshot(st *store.Store, agentCode string) (model.UsageSnapshot, error) {
	// Keep est_cost_usd fresh from the local pricing table (same call the Overview
	// uses). Cheap + idempotent; makes /api/usage correct on its own.
	if err := st.BackfillRollupCost(Cost); err != nil {
		return model.UsageSnapshot{}, err
	}

	today := todayDay()
	heatSince := dayInt(dayToTime(today).AddDate(0, 0, -(heatWindowDays - 1)))

	daily, err := st.UsageDaily(agentCode, heatSince)
	if err != nil {
		return model.UsageSnapshot{}, err
	}
	byDay := make(map[int]model.DailyUsage, len(daily))
	active := make(map[int]bool, len(daily))
	var heatMax int64
	for _, d := range daily {
		byDay[d.Day] = d
		if d.TotalTokens > 0 {
			active[d.Day] = true
		}
		if d.TotalTokens > heatMax {
			heatMax = d.TotalTokens
		}
	}

	snap := model.UsageSnapshot{WindowDays: usageWindowDays, Estimate: true}

	// 7-day trend, oldest→newest, zero-filled, weekday-labelled.
	for i := usageWindowDays - 1; i >= 0; i-- {
		t := dayToTime(today).AddDate(0, 0, -i)
		di := dayInt(t)
		d := byDay[di]
		d.Day = di
		d.Label = t.Format("Mon")
		snap.Days = append(snap.Days, d)
		snap.TotalTokens += d.TotalTokens
		snap.TotalCostUSD += d.CostUSD
	}

	// 56-day heatmap, oldest→newest.
	for i := heatWindowDays - 1; i >= 0; i-- {
		t := dayToTime(today).AddDate(0, 0, -i)
		di := dayInt(t)
		tot := byDay[di].TotalTokens
		snap.Heatmap = append(snap.Heatmap, model.HeatDay{
			Day: di, TotalTokens: tot, Bucket: heatBucket(tot, heatMax),
		})
	}

	snap.StreakDays = computeStreak(today, active)

	trendSince := dayInt(dayToTime(today).AddDate(0, 0, -(usageWindowDays - 1)))

	// By project (7-day window), labelled + percentaged against the top project.
	projects, err := st.UsageByProject(agentCode, trendSince)
	if err != nil {
		return model.UsageSnapshot{}, err
	}
	var maxCost float64
	for _, p := range projects {
		if p.CostUSD > maxCost {
			maxCost = p.CostUSD
		}
	}
	for i, p := range projects {
		if i >= maxProjects {
			break
		}
		pct := 0
		if maxCost > 0 {
			pct = int(p.CostUSD / maxCost * 100)
		}
		snap.ByProject = append(snap.ByProject, model.ProjectUsage{
			Name: projectLabel(p.Root), CostUSD: p.CostUSD, Pct: pct,
		})
	}

	// By model (7-day window), share of total tokens.
	models, err := st.UsageByModel(agentCode, trendSince)
	if err != nil {
		return model.UsageSnapshot{}, err
	}
	var totalToks int64
	for _, m := range models {
		totalToks += m.TotalTokens
	}
	for _, m := range models {
		pct := 0
		if totalToks > 0 {
			pct = int(float64(m.TotalTokens) / float64(totalToks) * 100)
		}
		snap.ByModel = append(snap.ByModel, model.UsageShare{Name: modelLabel(m.Model), Pct: pct})
	}
	sort.SliceStable(snap.ByModel, func(i, j int) bool { return snap.ByModel[i].Pct > snap.ByModel[j].Pct })

	return snap, nil
}

// QuotaSnapshot assembles the live plan-quota payload from the latest sample per
// window. When no samples exist the returned snapshot has Available=false and nil
// window pointers — the UI renders the "install the statusline helper" gated
// state. A store error short-circuits to a zero snapshot (failsafe, §14).
func QuotaSnapshot(st *store.Store, agentCode string) (model.QuotaSnapshot, error) {
	rows, err := st.LatestQuota(agentCode)
	if err != nil {
		return model.QuotaSnapshot{}, err
	}
	var snap model.QuotaSnapshot
	for _, r := range rows {
		w := &model.QuotaWindow{
			UsedPercentage: r.UsedPercentage,
			ResetsAtMs:     r.ResetsAtMs,
			TsMs:           r.TsMs,
		}
		switch r.Window {
		case "five_hour":
			snap.FiveHour = w
		case "seven_day":
			snap.SevenDay = w
		}
		// plan/source are the same across windows in practice; the last seen wins.
		if r.Plan != "" {
			snap.Plan = r.Plan
		}
		if r.Source != "" {
			snap.Source = r.Source
		}
	}
	// Available is true once at least one window has a sample.
	snap.Available = snap.FiveHour != nil || snap.SevenDay != nil
	return snap, nil
}
