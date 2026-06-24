package resolve

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

// agent constructs a minimal InventoryItem for agent-category tests.
// The "model" attr is set to "sonnet" so estTokens can be exercised.
func agent(name string, scope model.Scope) model.InventoryItem {
	return model.InventoryItem{
		AgentCode: "claude", Category: model.CatAgent,
		Name: name, Scope: scope, Enabled: true,
		Attrs: map[string]string{"model": "sonnet"},
	}
}

// TestResolveAgents_ProjectBeatsUser verifies the agents-specific precedence
// rule: project overrides user — the OPPOSITE of skills.
func TestResolveAgents_ProjectBeatsUser(t *testing.T) {
	items := []model.InventoryItem{agent("reviewer", model.ScopeUser), agent("reviewer", model.ScopeProject)}
	got := Resolve(items, model.Toggles{})
	r, ok := find(got, "reviewer")
	if !ok {
		t.Fatal("reviewer missing")
	}
	if r.EffectiveStatus != model.StatusActive || r.Winner.Scope != model.ScopeProject {
		t.Errorf("status/winner = %s/%s, want active/project", r.EffectiveStatus, r.Winner.Scope)
	}
	// Trail: weakest first, winner last.
	if len(r.PrecedenceTrail) != 2 {
		t.Fatalf("trail len = %d, want 2", len(r.PrecedenceTrail))
	}
	last := r.PrecedenceTrail[len(r.PrecedenceTrail)-1]
	if last.Decision != "wins" || last.Scope != string(model.ScopeProject) {
		t.Errorf("last step = {Decision:%s Scope:%s}, want {wins project}", last.Decision, last.Scope)
	}
	first := r.PrecedenceTrail[0]
	if first.Scope != string(model.ScopeUser) || first.Decision != "overridden" {
		t.Errorf("first step = {Decision:%s Scope:%s}, want {overridden user}", first.Decision, first.Scope)
	}
}
