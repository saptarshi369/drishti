package resolve

// commands.go — resolveCommands. Custom slash commands follow the SAME precedence
// as skills (enterprise > user > project). The cross-category rule: if a same-name
// skill is active, it takes precedence and the command resolves "shadowed".

import "github.com/saptarshi369/drishti/internal/model"

// commandOrder mirrors skillOrder: commands and skills share a precedence order.
var commandOrder = skillOrder

// resolveCommands resolves command precedence and applies skill shadowing.
// skillWinners holds the names of skills that resolved active; a command whose
// name is in that set is shadowed (the skill wins), regardless of scope.
func resolveCommands(items []model.InventoryItem, tg model.Toggles, skillWinners map[string]bool) []model.ResolvedItem {
	names, g := groupByName(items)
	var out []model.ResolvedItem
	for _, name := range names {
		group := g[name]
		win := pickByScopeOrder(group, commandOrder)
		winner := group[win]
		trail := buildTrail(group, win, commandOrder, "command")

		status := model.StatusActive
		if skillWinners[name] {
			status = model.StatusShadowed
			trail = append(trail, model.PrecedenceStep{
				Step: len(trail) + 1, Scope: "skill", Decision: "shadowed",
				Reason: "a same-name skill takes precedence over this command",
			})
		}

		res := model.ResolvedItem{
			AgentCode: winner.AgentCode, ProjectRoot: winner.ProjectRoot,
			Category: model.CatCommand, Name: name, EffectiveStatus: status,
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
