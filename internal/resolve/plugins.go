package resolve

// plugins.go — resolvePlugins. A plugin is active or disabled per the
// enabledPlugins map. When the same plugin is configured at multiple scopes, the
// highest-precedence scope (managed > local > project > user) decides the state.

import "github.com/saptarshi369/drishti/internal/model"

// pluginOrder is the plugin scope precedence, highest first.
var pluginOrder = []model.Scope{
	model.ScopeEnterprise, model.ScopeLocal, model.ScopeProject, model.ScopeUser,
}

// resolvePlugins picks the highest-precedence scope per plugin name; its Enabled
// flag becomes active (true) or disabled (false).
func resolvePlugins(items []model.InventoryItem) []model.ResolvedItem {
	names, g := groupByName(items)
	out := make([]model.ResolvedItem, 0, len(names))
	for _, name := range names {
		group := g[name]
		win := pickByScopeOrder(group, pluginOrder)
		winner := group[win]
		trail := buildTrail(group, win, pluginOrder, "plugin")

		status := model.StatusDisabled
		decision, reason := "disabled", "disabled in enabledPlugins"
		if winner.Enabled {
			status = model.StatusActive
			decision, reason = "wins", "enabled in enabledPlugins"
		}
		trail = append(trail, model.PrecedenceStep{
			Step: len(trail) + 1, Scope: string(winner.Scope), Decision: decision, Reason: reason,
		})

		res := model.ResolvedItem{
			AgentCode: winner.AgentCode, ProjectRoot: winner.ProjectRoot,
			Category: model.CatPlugin, Name: name, EffectiveStatus: status,
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
