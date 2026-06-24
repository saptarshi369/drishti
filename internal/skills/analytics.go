package skills

import "github.com/saptarshi369/drishti/internal/model"

// BuildAnalytics derives the Skills Analytics read-model from raw store rows.
// It is pure (no I/O) so it is trivially unit-testable from literal inputs.
//
// Per row it computes:
//   - value ratio = triggers / (est_context_tokens / 1000); 0 when tokens == 0
//     (e.g. disabled skills, which estimate 0) to avoid a divide-by-zero.
//   - dead            = active skill that has never fired (pure wasted context).
//   - over_triggering = active, costs tokens, fired >= HighTriggerMin times, yet
//     its value ratio is still below LowValueRatioMax (heavy but not earning it).
//   - disabled        = effective_status is "disabled" (turned off in settings).
//
// Items is always non-nil (made with len(rows) capacity) so the JSON shape is
// stable for the empty case ([] not null), like BuildSecuritySnapshot.
func BuildAnalytics(rows []model.SkillStatRow, t Thresholds) model.SkillsSnapshot {
	items := make([]model.SkillAnalyticsItem, 0, len(rows))
	var counts model.SkillCounts
	totalTokens := 0

	for _, r := range rows {
		active := r.EffectiveStatus == model.StatusActive
		disabled := r.EffectiveStatus == model.StatusDisabled

		// value ratio: triggers per 1k tokens. Guard the 0-token case (disabled
		// skills and any unestimated skill) so we never divide by zero.
		var ratio float64
		if r.EstContextTokens > 0 {
			ratio = float64(r.Triggers) / (float64(r.EstContextTokens) / 1000.0)
		}

		dead := active && r.Triggers == 0
		over := active && r.EstContextTokens > 0 &&
			r.Triggers >= t.HighTriggerMin && ratio < t.LowValueRatioMax

		items = append(items, model.SkillAnalyticsItem{
			Name:             r.Name,
			Triggers:         r.Triggers,
			LastFiredMs:      r.LastFiredMs,
			EstContextTokens: r.EstContextTokens,
			ValueRatio:       ratio,
			Dead:             dead,
			OverTriggering:   over,
			Disabled:         disabled,
		})

		if dead {
			counts.Dead++
		}
		if over {
			counts.OverTriggering++
		}
		if disabled {
			counts.Disabled++
		}
		totalTokens += r.EstContextTokens
	}

	return model.SkillsSnapshot{
		Items:              items,
		Counts:             counts,
		Total:              len(rows),
		TotalContextTokens: totalTokens,
	}
}
