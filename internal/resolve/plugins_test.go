package resolve

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func TestResolvePlugins(t *testing.T) {
	items := []model.InventoryItem{
		{Category: model.CatPlugin, Name: "github@official", Scope: model.ScopeUser, Enabled: true},
		{Category: model.CatPlugin, Name: "legacy@community", Scope: model.ScopeUser, Enabled: false},
		// Same plugin enabled at user but disabled at project; project wins state.
		{Category: model.CatPlugin, Name: "fmt@org", Scope: model.ScopeUser, Enabled: true},
		{Category: model.CatPlugin, Name: "fmt@org", Scope: model.ScopeProject, Enabled: false},
	}
	got := resolvePlugins(items)
	byName := map[string]model.ResolvedItem{}
	for _, r := range got {
		byName[r.Name] = r
	}
	if byName["github@official"].EffectiveStatus != model.StatusActive {
		t.Errorf("github status = %q", byName["github@official"].EffectiveStatus)
	}
	if byName["legacy@community"].EffectiveStatus != model.StatusDisabled {
		t.Errorf("legacy status = %q", byName["legacy@community"].EffectiveStatus)
	}
	if byName["fmt@org"].EffectiveStatus != model.StatusDisabled {
		t.Errorf("fmt status = %q, want disabled (project beats user)", byName["fmt@org"].EffectiveStatus)
	}
}
