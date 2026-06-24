package resolve

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

// TestBuildTrail_WinnerIsLast verifies that buildTrail always places the
// highest-precedence (winning) step last, regardless of input order.
// This is the core narrative property: the trail reads "weakest → strongest",
// ending in the decision that actually takes effect.
func TestBuildTrail_WinnerIsLast(t *testing.T) {
	// group has user first (higher precedence for skills), project second
	// (lower precedence). The winner is index 0 (user).
	group := []model.InventoryItem{
		{Name: "commit", Scope: model.ScopeUser, Category: model.CatSkill},
		{Name: "commit", Scope: model.ScopeProject, Category: model.CatSkill},
	}
	trail := buildTrail(group, 0, skillOrder, "skill")
	if len(trail) != 2 {
		t.Fatalf("len(trail) = %d, want 2", len(trail))
	}
	last := trail[len(trail)-1]
	if last.Decision != "wins" || last.Scope != string(model.ScopeUser) {
		t.Errorf("last step = {Decision:%s Scope:%s}, want {wins user}", last.Decision, last.Scope)
	}
	first := trail[0]
	if first.Decision != "overridden" || first.Scope != string(model.ScopeProject) {
		t.Errorf("first step = {Decision:%s Scope:%s}, want {overridden project}", first.Decision, first.Scope)
	}
}

func TestResolve_DispatchAllCategories(t *testing.T) {
	items := []model.InventoryItem{
		// A skill and a same-name command: the command must end up shadowed.
		{Category: model.CatSkill, Name: "deploy", Scope: model.ScopeUser, Attrs: map[string]string{"description": "s"}},
		{Category: model.CatCommand, Name: "deploy", Scope: model.ScopeProject, Attrs: map[string]string{}},
		{Category: model.CatMemory, Name: "CLAUDE.md (user)", Scope: model.ScopeUser, Attrs: map[string]string{"bytes": "40"}},
		{Category: model.CatPlugin, Name: "github@official", Scope: model.ScopeUser, Enabled: true},
	}
	got := Resolve(items, model.Toggles{})

	var cmd, mem, plugin, def *model.ResolvedItem
	for i := range got {
		switch {
		case got[i].Category == model.CatCommand && got[i].Name == "deploy":
			cmd = &got[i]
		case got[i].Category == model.CatMemory:
			mem = &got[i]
		case got[i].Category == model.CatPlugin:
			plugin = &got[i]
		case got[i].Category == model.CatOutputStyle && got[i].Name == "Default":
			def = &got[i]
		}
	}
	if cmd == nil || cmd.EffectiveStatus != model.StatusShadowed {
		t.Errorf("deploy command not shadowed: %+v", cmd)
	}
	if mem == nil || mem.EffectiveStatus != model.StatusActive {
		t.Errorf("memory not active: %+v", mem)
	}
	if plugin == nil || plugin.EffectiveStatus != model.StatusActive {
		t.Errorf("plugin not active: %+v", plugin)
	}
	if def == nil || def.EffectiveStatus != model.StatusActive {
		t.Errorf("Default output style should be active when unset: %+v", def)
	}
}
