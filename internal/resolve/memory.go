package resolve

// memory.go — resolveMemory. Like hooks, memory files from all scopes MERGE into
// context; there is no name contest, so each file becomes its own active row.
// The only suppression is claudeMdExcludes (absolute-path globs) → disabled.

import (
	"strconv"

	"github.com/saptarshi369/drishti/internal/model"
)

// resolveMemory marks every memory file active unless its absolute path matches a
// claudeMdExcludes glob. EstContextTokens comes from the stored byte count
// because memory files are real, launch-time context cost (feeds Module 4).
func resolveMemory(items []model.InventoryItem, tg model.Toggles) []model.ResolvedItem {
	out := make([]model.ResolvedItem, 0, len(items))
	for _, it := range items {
		// Copy so the Winner pointer owns stable memory (loop-variable hazard).
		w := it
		status := model.StatusActive
		trail := []model.PrecedenceStep{{
			Step: 1, Scope: string(it.Scope), Decision: "wins",
			Reason: "memory files from all scopes merge into context",
		}}
		if excludedByGlob(it.Attrs["abs"], tg.ClaudeMdExcludes) {
			status = model.StatusDisabled
			trail = append(trail, model.PrecedenceStep{
				Step: 2, Scope: "settings", Decision: "disabled",
				Reason: "excluded by claudeMdExcludes",
			})
		}
		res := model.ResolvedItem{
			AgentCode: it.AgentCode, ProjectRoot: it.ProjectRoot,
			Category: model.CatMemory, Name: it.Name, EffectiveStatus: status,
			PrecedenceTrail: trail,
		}
		if status == model.StatusActive {
			res.Winner = &w
		}
		out = append(out, res)
	}
	return out
}

// excludedByGlob reports whether path matches any of the claudeMdExcludes globs.
func excludedByGlob(path string, globs []string) bool {
	for _, g := range globs {
		if globMatch(g, path) {
			return true
		}
	}
	return false
}

// atoiSafe parses a base-10 int, returning 0 on any error (a missing or bad
// "bytes" attr just means a zero token estimate — never a failure, §14).
func atoiSafe(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
