package resolve

import "github.com/saptarshi369/drishti/internal/model"

// skillOrder is the documented skill precedence: managed > user (personal) >
// project > bundled. (Plugin skills are namespaced and cannot conflict.)
var skillOrder = []model.Scope{model.ScopeEnterprise, model.ScopeUser, model.ScopeProject, model.ScopeBundled}

// resolveSkills applies skill precedence + the disableBundledSkills /
// skillOverrides toggles. The winner is the highest-precedence scope; lower
// scopes of the same name are marked overridden. An "off" skillOverride or a
// disabled bundled skill yields a disabled (no-winner) result.
func resolveSkills(items []model.InventoryItem, tg model.Toggles) []model.ResolvedItem {
	names, g := groupByName(items)
	var out []model.ResolvedItem
	for _, name := range names {
		group := g[name]
		win := pickByScopeOrder(group, skillOrder)
		winner := group[win]

		// buildTrail sorts candidates weakest-first so the winner is always
		// emitted last, giving a consistent narrative regardless of input order.
		trail := buildTrail(group, win, skillOrder, "skill")

		status := model.StatusActive
		if override := tg.SkillOverrides[name]; override == "off" {
			status = model.StatusDisabled
			trail = append(trail, model.PrecedenceStep{Step: len(trail) + 1, Scope: "settings", Decision: "disabled", Reason: "skillOverrides set this skill off"})
		} else if winner.Scope == model.ScopeBundled && tg.DisableBundledSkills {
			status = model.StatusDisabled
			trail = append(trail, model.PrecedenceStep{Step: len(trail) + 1, Scope: "settings", Decision: "disabled", Reason: "disableBundledSkills hides bundled skills"})
		}

		res := model.ResolvedItem{
			AgentCode: winner.AgentCode, ProjectRoot: winner.ProjectRoot,
			Category: model.CatSkill, Name: name, EffectiveStatus: status,
			PrecedenceTrail: trail,
		}
		if status == model.StatusActive {
			w := winner
			res.Winner = &w
		}
		out = append(out, res)
	}
	return out
}

// scopeReason builds a human sentence for one trail step.
func scopeReason(scope model.Scope, decision, kind string) string {
	switch decision {
	case "wins":
		return string(scope) + " scope wins for this " + kind
	case "overridden":
		return string(scope) + " scope is overridden by a higher-precedence scope"
	default:
		return string(scope) + " scope defines this " + kind
	}
}
