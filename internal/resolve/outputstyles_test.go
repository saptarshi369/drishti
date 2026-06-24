package resolve

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func TestResolveOutputStyles_Selection(t *testing.T) {
	items := []model.InventoryItem{
		{Category: model.CatOutputStyle, Name: "Diagrams first", Scope: model.ScopeUser, Attrs: map[string]string{"description": "d"}},
	}
	// Selected style is "Explanatory" (a built-in).
	got := resolveOutputStyles(items, model.Toggles{OutputStyle: "Explanatory"})

	byName := map[string]model.ResolvedItem{}
	for _, r := range got {
		byName[r.Name] = r
	}
	// All 4 built-ins present + the custom one = 5 rows.
	if len(got) != 5 {
		t.Fatalf("want 5 rows, got %d", len(got))
	}
	if byName["Explanatory"].EffectiveStatus != model.StatusActive {
		t.Errorf("Explanatory status = %q, want active", byName["Explanatory"].EffectiveStatus)
	}
	if byName["Explanatory"].Winner == nil {
		t.Error("active style needs a winner")
	}
	if byName["Default"].EffectiveStatus != model.StatusDisabled {
		t.Errorf("Default status = %q, want disabled", byName["Default"].EffectiveStatus)
	}
	if byName["Diagrams first"].EffectiveStatus != model.StatusDisabled {
		t.Errorf("custom status = %q, want disabled", byName["Diagrams first"].EffectiveStatus)
	}
}

func TestResolveOutputStyles_DefaultWhenUnset(t *testing.T) {
	got := resolveOutputStyles(nil, model.Toggles{})
	for _, r := range got {
		if r.Name == "Default" && r.EffectiveStatus != model.StatusActive {
			t.Errorf("unset outputStyle should select Default, got %q", r.EffectiveStatus)
		}
	}
}

// TestResolveOutputStyles_CustomByName verifies a CUSTOM style is selected when
// the outputStyle setting matches its resolved Name (frontmatter name or slug),
// the same identifier Claude Code's outputStyle setting stores.
func TestResolveOutputStyles_CustomByName(t *testing.T) {
	items := []model.InventoryItem{
		{Category: model.CatOutputStyle, Name: "Diagrams first", Scope: model.ScopeUser, Attrs: map[string]string{"description": "d"}},
	}
	got := resolveOutputStyles(items, model.Toggles{OutputStyle: "Diagrams first"})
	byName := map[string]model.ResolvedItem{}
	for _, r := range got {
		byName[r.Name] = r
	}
	if byName["Diagrams first"].EffectiveStatus != model.StatusActive {
		t.Errorf("custom style status = %q, want active", byName["Diagrams first"].EffectiveStatus)
	}
	if byName["Default"].EffectiveStatus != model.StatusDisabled {
		t.Errorf("Default status = %q, want disabled (custom is selected)", byName["Default"].EffectiveStatus)
	}
}
