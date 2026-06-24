package resolve

// hooks.go — resolveHooks turns raw hook InventoryItems into ResolvedItems.
// Unlike skills/agents/mcp, hooks from all scopes MERGE and all run — there is
// no winner/loser contest. Each hook therefore becomes its own active resolved
// row. When two hooks share the same display name (e.g. the same event pattern
// defined in both user and project settings), the second occurrence is suffixed
// " (2)", the third " (3)", etc., so the resolved (category, name) key stays
// unique across the output slice.

import (
	"fmt"

	"github.com/saptarshi369/drishti/internal/model"
)

// resolveHooks treats every hook as active: hooks from all scopes MERGE and run
// (Claude deduplicates identical handlers at the source). There is no name
// contest, so each hook becomes its own resolved row. Display names that collide
// get a " (2)", " (3)" suffix so the resolved (category, name) key stays unique.
//
// The precedence trail for each hook has a single step explaining the merge
// policy — unlike buildTrail (which records a contest), this is informational.
func resolveHooks(items []model.InventoryItem) []model.ResolvedItem {
	// seen tracks how many times each display name has been emitted so far.
	// When a name appears for the first time (count == 1) it keeps its original
	// name; subsequent occurrences get " (n)" appended.
	seen := map[string]int{}

	var out []model.ResolvedItem
	for _, it := range items {
		// Increment before use: first occurrence becomes 1 (no suffix),
		// second becomes 2 → " (2)", and so on.
		seen[it.Name]++
		name := it.Name
		if n := seen[it.Name]; n > 1 {
			// Append the disambiguating counter so the resolved name is unique.
			name = fmt.Sprintf("%s (%d)", it.Name, n)
		}

		// Copy the item so the Winner pointer owns stable memory independent of
		// the loop variable (pre-Go-1.22 loop-variable aliasing hazard).
		w := it

		out = append(out, model.ResolvedItem{
			AgentCode:   it.AgentCode,
			ProjectRoot: it.ProjectRoot,
			Category:    model.CatHook,
			Name:        name,
			// Every hook is active — the merge policy means no hook is suppressed.
			EffectiveStatus: model.StatusActive,
			Winner:          &w,
			// Single informational step: explains WHY this hook is active.
			PrecedenceTrail: []model.PrecedenceStep{
				{
					Step:     1,
					Scope:    string(it.Scope),
					Decision: "wins",
					Reason:   "hooks from all scopes merge and run",
				},
			},
		})
	}
	return out
}
