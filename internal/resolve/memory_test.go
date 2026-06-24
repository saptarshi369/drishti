package resolve

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func TestResolveMemory(t *testing.T) {
	items := []model.InventoryItem{
		{Category: model.CatMemory, Name: "CLAUDE.md (user)", Scope: model.ScopeUser,
			Attrs: map[string]string{"bytes": "400", "abs": "/home/u/.claude/CLAUDE.md"}},
		{Category: model.CatMemory, Name: "other-team/CLAUDE.md (project)", Scope: model.ScopeProject,
			Attrs: map[string]string{"bytes": "40", "abs": "/repo/other-team/CLAUDE.md"}},
	}
	tg := model.Toggles{ClaudeMdExcludes: []string{"**/other-team/CLAUDE.md"}}
	got := resolveMemory(items, tg)
	if len(got) != 2 {
		t.Fatalf("want 2, got %d", len(got))
	}
	// First (user) is active.
	if got[0].EffectiveStatus != model.StatusActive {
		t.Errorf("user status = %q", got[0].EffectiveStatus)
	}
	if got[0].Winner == nil {
		t.Error("active memory must have a winner")
	}
	// Second is disabled by claudeMdExcludes; no winner.
	if got[1].EffectiveStatus != model.StatusDisabled {
		t.Errorf("excluded status = %q", got[1].EffectiveStatus)
	}
	if got[1].Winner != nil {
		t.Error("disabled memory must not have a winner")
	}
}
