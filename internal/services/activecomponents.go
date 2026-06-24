package services

import (
	"fmt"
	"sort"

	"github.com/saptarshi369/drishti/internal/model"
)

// activeHereOrder is the fixed display order for the "Active here" panel's
// category rows. Categories not listed still count toward ActiveComponents but
// get no summary row (keeps the panel focused on the high-signal categories).
var activeHereOrder = []string{"skill", "mcp", "hook", "agent", "command"}

// buildActiveInventory folds resolved rows into the Overview active-component
// census and the "Active here" summary rows. It counts only rows whose effective
// status is "active" (overridden/shadowed/disabled are not in effect), matching
// BuildContextBudget. Pure: rows in → snapshots out.
func buildActiveInventory(rows []model.ResolvedRow) (model.ActiveComponents, []model.ActiveHereRow) {
	type agg struct{ count, user, project int }
	byCat := map[string]*agg{}
	total := 0
	for _, r := range rows {
		if r.EffectiveStatus != "active" {
			continue
		}
		total++
		a := byCat[r.Category]
		if a == nil {
			a = &agg{}
			byCat[r.Category] = a
		}
		a.count++
		if r.InUser {
			a.user++
		}
		if r.InProject {
			a.project++
		}
	}

	// ActiveComponents.ByCategory: every active category, sorted by count desc
	// then name asc for deterministic output the tests can assert.
	comps := model.ActiveComponents{Total: total, ByCategory: []model.CategoryCount{}}
	for cat, a := range byCat {
		comps.ByCategory = append(comps.ByCategory, model.CategoryCount{
			Category: cat, Count: a.count, UserCount: a.user, ProjectCount: a.project,
		})
	}
	sort.Slice(comps.ByCategory, func(i, j int) bool {
		if comps.ByCategory[i].Count != comps.ByCategory[j].Count {
			return comps.ByCategory[i].Count > comps.ByCategory[j].Count
		}
		return comps.ByCategory[i].Category < comps.ByCategory[j].Category
	})

	// Active-here rows: only the major categories, in the fixed display order,
	// and only when present.
	here := []model.ActiveHereRow{}
	for _, cat := range activeHereOrder {
		a := byCat[cat]
		if a == nil {
			continue
		}
		here = append(here, model.ActiveHereRow{
			Category: cat, Count: a.count, Note: scopeNote(a.user, a.project), CTA: "inventory",
		})
	}
	return comps, here
}

// scopeNote renders the user/project split for an Active-here row, omitting a
// zero side. Returns "" when both are zero.
func scopeNote(user, project int) string {
	switch {
	case user > 0 && project > 0:
		return fmt.Sprintf("%d user · %d project", user, project)
	case user > 0:
		return fmt.Sprintf("%d user", user)
	case project > 0:
		return fmt.Sprintf("%d project", project)
	default:
		return ""
	}
}
