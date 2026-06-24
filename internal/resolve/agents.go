package resolve

// agents.go — implements the subagent precedence resolver.
// Agents follow project > user order (opposite of skills): a project-scoped
// definition always wins over a user-scoped one. Enterprise is strongest, then
// project, then user, then plugin.

import "github.com/saptarshi369/drishti/internal/model"

// agentOrder is the documented subagent precedence, highest first:
// enterprise > project > user > plugin.
// (CLI --agents and nearest-cwd runtime selection are out of scope here.)
var agentOrder = []model.Scope{
	model.ScopeEnterprise,
	model.ScopeProject,
	model.ScopeUser,
	model.ScopePlugin,
}

// resolveAgents applies subagent precedence across all scopes: the
// highest-precedence scope's definition wins for each unique agent name.
// Unlike skills, agents have no disable toggles in this category; every
// resolved agent is always StatusActive.
func resolveAgents(items []model.InventoryItem) []model.ResolvedItem {
	// groupByName buckets same-name items together, returning a stable sorted
	// name list so output order is deterministic across runs.
	names, groups := groupByName(items)

	out := make([]model.ResolvedItem, 0, len(names))
	for _, name := range names {
		group := groups[name]

		// pickByScopeOrder returns the index of the strongest-scope item per
		// agentOrder. That index becomes the winner.
		win := pickByScopeOrder(group, agentOrder)
		winner := group[win]

		// buildTrail produces the ordered "why?" narrative: weakest scope first,
		// winner last with decision "wins". Shared helper — no inline loop.
		trail := buildTrail(group, win, agentOrder, "agent")

		// Copy winner to obtain an addressable pointer (loop variable is a copy).
		w := winner
		out = append(out, model.ResolvedItem{
			AgentCode:       winner.AgentCode,
			ProjectRoot:     winner.ProjectRoot,
			Category:        model.CatAgent,
			Name:            name,
			EffectiveStatus: model.StatusActive,
			Winner:          &w,
			PrecedenceTrail: trail,
		})
	}
	return out
}
