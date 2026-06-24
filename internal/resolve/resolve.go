// Package resolve turns raw, per-scope InventoryItems into the materialized
// "Effective" view plus an ordered "why?" precedence trail. Precedence is
// CATEGORY-SPECIFIC (skills: user>project; agents/mcp: project>user; hooks
// merge), matching Claude Code's documented behavior — there is no single global
// order. Every resolver is a PURE function so it is exhaustively table-testable;
// that table is the correctness proof the owner cannot get from reading Go.
package resolve

import (
	"sort"

	"github.com/saptarshi369/drishti/internal/model"
)

// Resolve dispatches each category to its own resolver and concatenates the
// results. Items may span multiple categories and scopes.
func Resolve(items []model.InventoryItem, tg model.Toggles) []model.ResolvedItem {
	byCat := map[model.Category][]model.InventoryItem{}
	for _, it := range items {
		byCat[it.Category] = append(byCat[it.Category], it)
	}
	var out []model.ResolvedItem

	// Skills must resolve before commands: a command is shadowed by a same-name
	// ACTIVE skill, so we collect the winning skill names first.
	skillResolved := resolveSkills(byCat[model.CatSkill], tg)
	out = append(out, skillResolved...)
	skillWinners := map[string]bool{}
	for _, r := range skillResolved {
		if r.EffectiveStatus == model.StatusActive {
			skillWinners[r.Name] = true
		}
	}

	out = append(out, resolveAgents(byCat[model.CatAgent])...)
	out = append(out, resolveMCP(byCat[model.CatMCP], tg)...)
	out = append(out, resolveHooks(byCat[model.CatHook])...)
	out = append(out, resolveMemory(byCat[model.CatMemory], tg)...)
	out = append(out, resolveCommands(byCat[model.CatCommand], tg, skillWinners)...)
	out = append(out, resolveOutputStyles(byCat[model.CatOutputStyle], tg)...)
	out = append(out, resolvePlugins(byCat[model.CatPlugin])...)
	// Module 4: fill est_context_tokens in one consolidated, pluggable place
	// rather than inline per resolver. The default heuristic estimator charges
	// memory by bytes, skills/agents/commands by their listing line, MCP a flat
	// per-server constant, and everything non-active 0.
	return EstimateAll(out, DefaultEstimator())
}

// estTokens is the rough chars/4 context-cost heuristic feeding Module 4. It is
// deliberately crude; real analysis is the Context-Budget module's job.
func estTokens(text string) int { return estTokensN(len(text)) }

// estTokensN is estTokens over a precomputed character count (used by memory,
// which stores only the byte size in Attrs rather than the full text).
func estTokensN(n int) int { return (n + 3) / 4 }

// buildTrail returns precedence-ordered trail steps for one name group:
// candidates are listed from lowest to highest precedence (so the winner is
// last), the winner's step is "wins" and the rest are "overridden". This makes
// the "why?" trail read as a narrative that ends in the decision, independent
// of the order items were discovered in. winIdx is the index in group of the
// winning item; order is the category's scope precedence (highest first).
func buildTrail(group []model.InventoryItem, winIdx int, order []model.Scope, kind string) []model.PrecedenceStep {
	// rank maps a scope to its precedence index (lower = stronger). Unknown
	// scopes sort after all known ones so they appear first (weakest).
	rank := func(s model.Scope) int {
		for i, sc := range order {
			if sc == s {
				return i
			}
		}
		return len(order)
	}
	// Indices into group, sorted by DESCENDING strength (weakest first → winner last).
	idx := make([]int, len(group))
	for i := range group {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool {
		return rank(group[idx[a]].Scope) > rank(group[idx[b]].Scope)
	})
	trail := make([]model.PrecedenceStep, 0, len(group))
	for _, gi := range idx {
		dec := "overridden"
		if gi == winIdx {
			dec = "wins"
		}
		trail = append(trail, model.PrecedenceStep{
			Step: len(trail) + 1, Scope: string(group[gi].Scope),
			Decision: dec, Reason: scopeReason(group[gi].Scope, dec, kind),
		})
	}
	return trail
}

// groupByName buckets items by Name, preserving input order within a bucket.
func groupByName(items []model.InventoryItem) ([]string, map[string][]model.InventoryItem) {
	g := map[string][]model.InventoryItem{}
	var names []string
	for _, it := range items {
		if _, seen := g[it.Name]; !seen {
			names = append(names, it.Name)
		}
		g[it.Name] = append(g[it.Name], it)
	}
	sort.Strings(names)
	return names, g
}

// pickByScopeOrder returns the index in items of the highest-precedence scope
// found, given order (earliest = highest). Returns -1 if none match.
func pickByScopeOrder(items []model.InventoryItem, order []model.Scope) int {
	best, bestRank := -1, len(order)
	for i, it := range items {
		for rank, sc := range order {
			if it.Scope == sc && rank < bestRank {
				best, bestRank = i, rank
			}
		}
	}
	return best
}
