package resolve

// outputstyles.go — resolveOutputStyles. Output styles are a PICK-ONE selection,
// not a merge: exactly one style (the one named by the outputStyle setting, or
// "Default") is active. Same-name styles across scopes resolve by precedence
// (project/closest wins; the loser shows as overridden in the trail); every
// surviving style that is NOT the selected one renders disabled ("present, not
// in effect"). The 4 built-ins are synthesized so they always appear.

import "github.com/saptarshi369/drishti/internal/model"

// outputStyleOrder: managed > project > user > bundled (project beats user, like
// agents — the closest-to-cwd definition wins a name clash).
var outputStyleOrder = []model.Scope{
	model.ScopeEnterprise, model.ScopeProject, model.ScopeUser, model.ScopeBundled,
}

// builtinOutputStyles are always available even with no custom files on disk.
var builtinOutputStyles = []string{"Default", "Proactive", "Explanatory", "Learning"}

// resolveOutputStyles injects the built-ins, resolves same-name scope clashes,
// then marks the single selected style active and the rest disabled.
func resolveOutputStyles(items []model.InventoryItem, tg model.Toggles) []model.ResolvedItem {
	// Start from the discovered items, then add any built-in not already defined
	// by a custom file of the same name (a custom file overrides the built-in).
	all := append([]model.InventoryItem{}, items...)
	have := map[string]bool{}
	for _, it := range items {
		have[it.Name] = true
	}
	for _, b := range builtinOutputStyles {
		if !have[b] {
			all = append(all, model.InventoryItem{
				AgentCode: "claude", Category: model.CatOutputStyle, Name: b,
				Scope: model.ScopeBundled, RelPath: "(built-in)", Enabled: true,
				Attrs: map[string]string{"description": "built-in output style"},
			})
		}
	}

	// "" means no outputStyle set, which Claude treats as the built-in "Default".
	selected := tg.OutputStyle
	if selected == "" {
		selected = "Default"
	}

	names, g := groupByName(all)
	out := make([]model.ResolvedItem, 0, len(names))
	for _, name := range names {
		group := g[name]
		win := pickByScopeOrder(group, outputStyleOrder)
		winner := group[win]
		trail := buildTrail(group, win, outputStyleOrder, "output style")

		status := model.StatusDisabled
		if name == selected {
			status = model.StatusActive
			trail = append(trail, model.PrecedenceStep{
				Step: len(trail) + 1, Scope: "settings", Decision: "wins",
				Reason: "selected via outputStyle setting",
			})
		} else {
			trail = append(trail, model.PrecedenceStep{
				Step: len(trail) + 1, Scope: "settings", Decision: "disabled",
				Reason: "not selected (outputStyle = " + selected + ")",
			})
		}

		res := model.ResolvedItem{
			AgentCode: winner.AgentCode, ProjectRoot: winner.ProjectRoot,
			Category: model.CatOutputStyle, Name: name, EffectiveStatus: status,
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
