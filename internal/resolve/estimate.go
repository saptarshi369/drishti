package resolve

// estimate.go — the pluggable context-token estimator (Module 4). All per-category
// "always-on context tax" estimation lives here, consolidated out of the individual
// resolvers so the rules sit in one tested place and alternate strategies (e.g. a
// future `claude mcp list`-backed MCP estimator) can drop in without touching callers.

import "github.com/saptarshi369/drishti/internal/model"

// Estimator estimates the always-on context cost, in tokens, that a resolved
// inventory item contributes to the system prompt before the user prompts.
// Implementations must be pure. A non-active item always estimates 0.
type Estimator interface {
	Estimate(r model.ResolvedItem) int
}

// HeuristicEstimator is the default character-count-based estimator. It is
// deliberately crude (chars/4) but consistent and honest per category; MCP tool
// definitions cannot be sized from config alone, so they use a flat per-server
// constant (real introspection via `claude mcp list` is a deferred follow-on).
type HeuristicEstimator struct {
	// MCPTokensPerServer is the flat estimate charged to each active MCP server
	// as a stand-in for its (unmeasured) tool-definition footprint.
	MCPTokensPerServer int
}

// DefaultEstimator returns the standard heuristic estimator used in production.
func DefaultEstimator() HeuristicEstimator {
	return HeuristicEstimator{MCPTokensPerServer: 500}
}

// Estimate implements Estimator. Only active items with a winner contribute; the
// basis is category-specific (see the Module 4 spec §2):
//   - memory:        full file bytes / 4 (injected verbatim)
//   - skill/agent/command: (name + description) / 4 (the system-prompt listing line)
//   - output_style:  description / 4 (the selected style)
//   - mcp:           a flat per-server constant
//   - hook/plugin:   0 (not injected context / counted via their own items)
func (e HeuristicEstimator) Estimate(r model.ResolvedItem) int {
	if r.EffectiveStatus != model.StatusActive || r.Winner == nil {
		return 0
	}
	switch r.Category {
	case model.CatMemory:
		return estTokensN(atoiSafe(r.Winner.Attrs["bytes"]))
	case model.CatSkill, model.CatAgent, model.CatCommand:
		return estTokens(r.Name + r.Winner.Attrs["description"])
	case model.CatOutputStyle:
		return estTokens(r.Winner.Attrs["description"])
	case model.CatMCP:
		return e.MCPTokensPerServer
	default: // hook, plugin
		return 0
	}
}

// EstimateAll fills EstContextTokens on every item using est, in place, and
// returns the same slice so callers can chain it onto a Resolve result.
func EstimateAll(items []model.ResolvedItem, est Estimator) []model.ResolvedItem {
	for i := range items {
		items[i].EstContextTokens = est.Estimate(items[i])
	}
	return items
}
