package resolve

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func skill(name string, scope model.Scope) model.InventoryItem {
	return model.InventoryItem{AgentCode: "claude", Category: model.CatSkill, Name: name, Scope: scope, Enabled: true, Attrs: map[string]string{"description": "x"}}
}

func find(rs []model.ResolvedItem, name string) (model.ResolvedItem, bool) {
	for _, r := range rs {
		if r.Name == name {
			return r, true
		}
	}
	return model.ResolvedItem{}, false
}

func TestResolveSkills_UserBeatsProject(t *testing.T) {
	// Skills: user (personal) overrides project — the OPPOSITE of agents/mcp.
	items := []model.InventoryItem{skill("commit", model.ScopeProject), skill("commit", model.ScopeUser)}
	got := Resolve(items, model.Toggles{})
	r, ok := find(got, "commit")
	if !ok {
		t.Fatal("commit missing")
	}
	if r.EffectiveStatus != model.StatusActive || r.Winner.Scope != model.ScopeUser {
		t.Errorf("status/winner = %s/%s, want active/user", r.EffectiveStatus, r.Winner.Scope)
	}
	if len(r.PrecedenceTrail) == 0 || r.PrecedenceTrail[len(r.PrecedenceTrail)-1].Decision != "wins" {
		t.Errorf("trail = %+v", r.PrecedenceTrail)
	}

	// Strengthen: both candidates must appear and the trail must be ordered
	// weakest-first (project overridden, user wins last).
	if len(r.PrecedenceTrail) != 2 {
		t.Fatalf("trail len = %d, want 2", len(r.PrecedenceTrail))
	}
	last := r.PrecedenceTrail[len(r.PrecedenceTrail)-1]
	if last.Decision != "wins" || last.Scope != string(model.ScopeUser) {
		t.Errorf("last step = {Decision:%s Scope:%s}, want {wins user}", last.Decision, last.Scope)
	}
	first := r.PrecedenceTrail[0]
	if first.Scope != string(model.ScopeProject) || first.Decision != "overridden" {
		t.Errorf("first step = {Decision:%s Scope:%s}, want {overridden project}", first.Decision, first.Scope)
	}
}

// TestResolveSkills_TrailWinsLastRegardlessOfInputOrder proves that the trail
// ordering is deterministic and does NOT depend on the order items were
// discovered. When user and project both define the same skill, user wins —
// and that winning step must be LAST in the trail regardless of which scope
// appeared first in the input slice.
func TestResolveSkills_TrailWinsLastRegardlessOfInputOrder(t *testing.T) {
	// Input order reversed vs TestResolveSkills_UserBeatsProject: user FIRST.
	items := []model.InventoryItem{skill("commit", model.ScopeUser), skill("commit", model.ScopeProject)}
	got := Resolve(items, model.Toggles{})
	r, ok := find(got, "commit")
	if !ok {
		t.Fatal("commit missing")
	}
	if len(r.PrecedenceTrail) == 0 {
		t.Fatal("empty trail")
	}
	last := r.PrecedenceTrail[len(r.PrecedenceTrail)-1]
	if last.Decision != "wins" || last.Scope != string(model.ScopeUser) {
		t.Errorf("last step = {Decision:%s Scope:%s}, want {wins user} — trail is input-order-dependent", last.Decision, last.Scope)
	}
}

func TestResolveSkills_OverrideOff(t *testing.T) {
	items := []model.InventoryItem{skill("pdf", model.ScopeUser)}
	got := Resolve(items, model.Toggles{SkillOverrides: map[string]string{"pdf": "off"}})
	r, _ := find(got, "pdf")
	if r.EffectiveStatus != model.StatusDisabled {
		t.Errorf("status = %s, want disabled", r.EffectiveStatus)
	}
}

func TestResolveSkills_DisableBundled(t *testing.T) {
	items := []model.InventoryItem{skill("explain", model.ScopeBundled)}
	got := Resolve(items, model.Toggles{DisableBundledSkills: true})
	r, _ := find(got, "explain")
	if r.EffectiveStatus != model.StatusDisabled {
		t.Errorf("status = %s, want disabled", r.EffectiveStatus)
	}
}
