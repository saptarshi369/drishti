package resolve

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func TestResolveCommands_ShadowAndPrecedence(t *testing.T) {
	items := []model.InventoryItem{
		// "deploy" exists as a command but a same-name active skill shadows it.
		{Category: model.CatCommand, Name: "deploy", Scope: model.ScopeProject, Attrs: map[string]string{}},
		// "lint" exists at user and project; user beats project (skill order).
		{Category: model.CatCommand, Name: "lint", Scope: model.ScopeProject, Attrs: map[string]string{"description": "proj"}},
		{Category: model.CatCommand, Name: "lint", Scope: model.ScopeUser, Attrs: map[string]string{"description": "user"}},
	}
	skillWinners := map[string]bool{"deploy": true}
	got := resolveCommands(items, model.Toggles{}, skillWinners)

	byName := map[string]model.ResolvedItem{}
	for _, r := range got {
		byName[r.Name] = r
	}
	if byName["deploy"].EffectiveStatus != model.StatusShadowed {
		t.Errorf("deploy status = %q, want shadowed", byName["deploy"].EffectiveStatus)
	}
	if byName["deploy"].Winner != nil {
		t.Error("shadowed command must not have a winner")
	}
	if byName["lint"].EffectiveStatus != model.StatusActive {
		t.Errorf("lint status = %q", byName["lint"].EffectiveStatus)
	}
	if byName["lint"].Winner == nil || byName["lint"].Winner.Scope != model.ScopeUser {
		t.Errorf("lint winner scope = %v, want user", byName["lint"].Winner)
	}
}

// TestResolveCommands_DisabledSkillDoesNotShadow verifies a command is shadowed
// ONLY by an ACTIVE same-name skill. A disabled skill is absent from skillWinners,
// so the command takes effect (active) — the subtle rule from the spec §2/§4.
func TestResolveCommands_DisabledSkillDoesNotShadow(t *testing.T) {
	items := []model.InventoryItem{
		{Category: model.CatCommand, Name: "deploy", Scope: model.ScopeProject, Attrs: map[string]string{}},
	}
	// "deploy" skill exists but is disabled (e.g. skillOverrides:off), so it is
	// NOT in skillWinners. The command must therefore be active, not shadowed.
	got := resolveCommands(items, model.Toggles{}, map[string]bool{})
	if len(got) != 1 {
		t.Fatalf("want 1, got %d", len(got))
	}
	if got[0].EffectiveStatus != model.StatusActive {
		t.Errorf("status = %q, want active (disabled skill must not shadow)", got[0].EffectiveStatus)
	}
}
