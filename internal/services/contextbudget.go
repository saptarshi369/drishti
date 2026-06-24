package services

// contextbudget.go — Module 4. Folds the resolved inventory rows into the
// Context-Budget snapshot: the always-on context "tax", a per-category breakdown,
// and the full list of active consumers (so the browser can recompute any
// "if disabled" combination client-side). Pure: rows in → snapshot out.

import (
	"sort"

	"github.com/saptarshi369/drishti/internal/model"
)

// BuildContextBudget assembles the Context-Budget snapshot from resolved rows.
// Only active rows count toward the tax (overridden/shadowed/disabled are not in
// effect, so they cost 0 tokens). windowTokens is the percentage denominator.
func BuildContextBudget(rows []model.ResolvedRow, windowTokens int) model.ContextBudgetSnapshot {
	snap := model.ContextBudgetSnapshot{
		WindowTokens: windowTokens,
		ByCategory:   []model.CategoryBudget{},
		Consumers:    []model.ConsumerItem{},
		Caveats:      []string{},
	}

	catTokens := map[string]int{}
	catCount := map[string]int{}
	hasMCP := false

	for _, r := range rows {
		if r.EffectiveStatus != "active" {
			continue
		}
		snap.TotalTokens += r.EstContextTokens
		catTokens[r.Category] += r.EstContextTokens
		catCount[r.Category]++
		if r.Category == "mcp" {
			hasMCP = true
		}
		snap.Consumers = append(snap.Consumers, model.ConsumerItem{
			ID: r.ID, Category: r.Category, Name: r.Name,
			Scope: r.WinnerScope, Tokens: r.EstContextTokens,
		})
	}

	// Per-category buckets, sorted by tokens desc (ties broken by name for stable
	// output the tests can assert).
	for cat, tok := range catTokens {
		snap.ByCategory = append(snap.ByCategory, model.CategoryBudget{
			Category: cat, Tokens: tok, Count: catCount[cat], Pct: pctOf(tok, snap.TotalTokens),
		})
	}
	sort.Slice(snap.ByCategory, func(i, j int) bool {
		if snap.ByCategory[i].Tokens != snap.ByCategory[j].Tokens {
			return snap.ByCategory[i].Tokens > snap.ByCategory[j].Tokens
		}
		return snap.ByCategory[i].Category < snap.ByCategory[j].Category
	})

	// Consumers sorted by tokens desc (ties by name) — the biggest-consumers table.
	sort.Slice(snap.Consumers, func(i, j int) bool {
		if snap.Consumers[i].Tokens != snap.Consumers[j].Tokens {
			return snap.Consumers[i].Tokens > snap.Consumers[j].Tokens
		}
		return snap.Consumers[i].Name < snap.Consumers[j].Name
	})

	snap.Pct = pctOf(snap.TotalTokens, windowTokens)

	if hasMCP {
		snap.Caveats = append(snap.Caveats,
			"MCP tool-definition cost is estimated at a flat ~500 tokens/server (not yet introspected via `claude mcp list`).")
	}
	return snap
}

// pctOf returns part/whole as a percentage, or 0 when whole is non-positive
// (avoids divide-by-zero; an unset window simply yields 0%).
func pctOf(part, whole int) float64 {
	if whole <= 0 {
		return 0
	}
	return float64(part) / float64(whole) * 100
}
