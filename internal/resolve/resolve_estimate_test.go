package resolve

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

// TestResolve_EstimatesAllCategories proves estimation flows through Resolve:
// the agent estimate now derives from description (not the model name), and the
// MCP server gets the flat constant.
func TestResolve_EstimatesAllCategories(t *testing.T) {
	items := []model.InventoryItem{
		{Category: model.CatAgent, Name: "rev", Scope: model.ScopeUser,
			Attrs: map[string]string{"model": "sonnet", "description": "Reviews code"}},
		{Category: model.CatMCP, Name: "github", Scope: model.ScopeUser,
			Attrs: map[string]string{"transport": "stdio"}},
	}
	got := Resolve(items, model.Toggles{})
	by := map[model.Category]int{}
	for _, r := range got {
		by[r.Category] = r.EstContextTokens
	}
	// agent: (len("rev")+len("Reviews code"))/4 = (3+12+3)/4 = 4 — NOT chars/4 of "sonnet".
	if by[model.CatAgent] != 4 {
		t.Errorf("agent est = %d, want 4 (from description, not model name)", by[model.CatAgent])
	}
	// mcp: the flat per-server constant.
	if by[model.CatMCP] != 500 {
		t.Errorf("mcp est = %d, want 500", by[model.CatMCP])
	}
}
