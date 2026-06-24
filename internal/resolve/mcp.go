package resolve

// mcp.go — MCP resolver: turns raw per-scope MCP server items into a single
// ResolvedItem per server name, applying precedence and toggle rules.

import "github.com/saptarshi369/drishti/internal/model"

// mcpOrder is the documented MCP server scope precedence: local beats project
// beats user beats plugin. This matches Claude Code's published behavior for
// mcp.json server definitions. (Skills follow the inverse order: user>project.)
var mcpOrder = []model.Scope{model.ScopeLocal, model.ScopeProject, model.ScopeUser, model.ScopePlugin}

// resolveMCP applies MCP precedence (whole-entry winner) and json-server
// toggles. A name in DisabledMcpjsonServers is always disabled. A non-nil
// EnabledMcpjsonServers acts as an allowlist: servers absent from it are also
// disabled. Status is config-only in Module 1 (no live connection probe is
// performed here — that is Module 5's job).
func resolveMCP(items []model.InventoryItem, tg model.Toggles) []model.ResolvedItem {
	// Build a quick lookup set for the deny list. Using a map[string]bool keeps
	// the inner loop O(1) rather than O(n*m).
	denied := map[string]bool{}
	for _, n := range tg.DisabledMcpjsonServers {
		denied[n] = true
	}

	// allow is nil when EnabledMcpjsonServers is nil (feature off). When it is
	// non-nil it becomes the allowlist: only servers in the map are permitted.
	var allow map[string]bool
	if tg.EnabledMcpjsonServers != nil {
		allow = make(map[string]bool, len(tg.EnabledMcpjsonServers))
		for _, n := range tg.EnabledMcpjsonServers {
			allow[n] = true
		}
	}

	// groupByName buckets all items by their Name field and returns a
	// deterministically ordered name slice (alphabetical) so output is stable.
	names, g := groupByName(items)

	out := make([]model.ResolvedItem, 0, len(names))
	for _, name := range names {
		group := g[name]

		// pickByScopeOrder finds the index of the highest-precedence item
		// (earliest in mcpOrder). For MCP servers this means local > project >
		// user > plugin (whole-entry semantics: the winning scope "wins" the
		// entire server definition, not per-field merge).
		win := pickByScopeOrder(group, mcpOrder)
		winner := group[win]

		// buildTrail returns steps ordered weakest-first, winner last, so the
		// "why?" explanation reads as a narrative ending in the decision.
		trail := buildTrail(group, win, mcpOrder, "MCP server")

		// Apply toggle rules AFTER the precedence winner is established. These
		// are settings-level overrides that can disable any server regardless of
		// scope.
		status := model.StatusActive
		switch {
		case denied[name]:
			// Explicitly listed in disabledMcpjsonServers.
			status = model.StatusDisabled
			trail = append(trail, model.PrecedenceStep{
				Step:     len(trail) + 1,
				Scope:    "settings",
				Decision: "disabled",
				Reason:   "listed in disabledMcpjsonServers",
			})
		case allow != nil && !allow[name]:
			// enabledMcpjsonServers is non-nil (allowlist mode) but this server
			// is not in it.
			status = model.StatusDisabled
			trail = append(trail, model.PrecedenceStep{
				Step:     len(trail) + 1,
				Scope:    "settings",
				Decision: "disabled",
				Reason:   "not in enabledMcpjsonServers allowlist",
			})
		}

		res := model.ResolvedItem{
			AgentCode:       winner.AgentCode,
			ProjectRoot:     winner.ProjectRoot,
			Category:        model.CatMCP,
			Name:            name,
			EffectiveStatus: status,
			PrecedenceTrail: trail,
		}
		// Winner is only populated when the server is active; a disabled server
		// has no "effective" definition to point at.
		if status == model.StatusActive {
			w := winner
			res.Winner = &w
		}
		out = append(out, res)
	}
	return out
}
