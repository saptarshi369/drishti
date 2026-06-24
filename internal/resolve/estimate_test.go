package resolve

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

// active builds an active ResolvedItem of a category with a winner carrying attrs.
func active(cat model.Category, name string, attrs map[string]string) model.ResolvedItem {
	w := model.InventoryItem{Category: cat, Name: name, Attrs: attrs}
	return model.ResolvedItem{Category: cat, Name: name, EffectiveStatus: model.StatusActive, Winner: &w}
}

func TestHeuristicEstimator_PerCategory(t *testing.T) {
	est := DefaultEstimator()
	cases := []struct {
		label string
		item  model.ResolvedItem
		want  int
	}{
		// memory: bytes/4 from the stored byte count → 400/4 = 100.
		{"memory", active(model.CatMemory, "CLAUDE.md", map[string]string{"bytes": "400"}), 100},
		// skill: (name+description)/4 → len("lint")+len("Runs the linter")=4+15=19 → (19+3)/4 = 5.
		{"skill", active(model.CatSkill, "lint", map[string]string{"description": "Runs the linter"}), 5},
		// agent: same basis as skill, from the NEW description attr (not the model name).
		{"agent", active(model.CatAgent, "rev", map[string]string{"description": "Reviews code", "model": "sonnet"}), 4},
		// command: (name+description)/4.
		{"command", active(model.CatCommand, "deploy", map[string]string{"description": "ship it"}), 4},
		// output-style: description-based.
		{"output_style", active(model.CatOutputStyle, "Default", map[string]string{"description": "built-in output style"}), 6},
		// mcp: flat per-server constant.
		{"mcp", active(model.CatMCP, "github", map[string]string{"transport": "stdio"}), 500},
		// hook/plugin: excluded → 0.
		{"hook", active(model.CatHook, "PreToolUse", nil), 0},
		{"plugin", active(model.CatPlugin, "snip", nil), 0},
	}
	for _, c := range cases {
		if got := est.Estimate(c.item); got != c.want {
			t.Errorf("%s: estimate = %d, want %d", c.label, got, c.want)
		}
	}
}

func TestHeuristicEstimator_NonActiveIsZero(t *testing.T) {
	est := DefaultEstimator()
	for _, st := range []model.EffectiveStatus{model.StatusDisabled, model.StatusShadowed, model.StatusOverridden} {
		// Even a memory file with real bytes contributes nothing when not active.
		w := model.InventoryItem{Category: model.CatMemory, Attrs: map[string]string{"bytes": "8000"}}
		item := model.ResolvedItem{Category: model.CatMemory, EffectiveStatus: st, Winner: &w}
		if got := est.Estimate(item); got != 0 {
			t.Errorf("status %s: estimate = %d, want 0", st, got)
		}
	}
}

func TestEstimateAll_SetsTokensInPlace(t *testing.T) {
	items := []model.ResolvedItem{
		active(model.CatMemory, "CLAUDE.md", map[string]string{"bytes": "400"}),
		active(model.CatMCP, "github", nil),
	}
	out := EstimateAll(items, DefaultEstimator())
	if out[0].EstContextTokens != 100 || out[1].EstContextTokens != 500 {
		t.Fatalf("EstimateAll tokens = %d,%d want 100,500", out[0].EstContextTokens, out[1].EstContextTokens)
	}
}
