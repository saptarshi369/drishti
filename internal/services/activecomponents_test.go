package services

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func TestBuildActiveInventoryCountsActiveOnly(t *testing.T) {
	rows := []model.ResolvedRow{
		{Category: "skill", Name: "a", EffectiveStatus: "active", InUser: true},
		{Category: "skill", Name: "b", EffectiveStatus: "active", InProject: true},
		{Category: "skill", Name: "c", EffectiveStatus: "disabled", InUser: true}, // excluded
		{Category: "mcp", Name: "m", EffectiveStatus: "active", InUser: true, InProject: true},
		{Category: "hook", Name: "h", EffectiveStatus: "shadowed"}, // excluded
	}
	comps, here := buildActiveInventory(rows)

	if comps.Total != 3 {
		t.Fatalf("total = %d, want 3 (active only)", comps.Total)
	}
	// Sorted by count desc: skill(2) before mcp(1).
	if comps.ByCategory[0].Category != "skill" || comps.ByCategory[0].Count != 2 {
		t.Errorf("byCat[0] = %+v, want skill/2", comps.ByCategory[0])
	}
	if comps.ByCategory[0].UserCount != 1 || comps.ByCategory[0].ProjectCount != 1 {
		t.Errorf("skill scope = %d/%d, want 1/1", comps.ByCategory[0].UserCount, comps.ByCategory[0].ProjectCount)
	}
	if comps.ByCategory[1].Category != "mcp" {
		t.Errorf("byCat[1] = %+v, want mcp", comps.ByCategory[1])
	}
	// Active-here: skill then mcp (fixed display order); hook excluded (shadowed).
	if len(here) != 2 || here[0].Category != "skill" || here[1].Category != "mcp" {
		t.Fatalf("here = %+v, want [skill, mcp]", here)
	}
	if here[0].Note != "1 user · 1 project" {
		t.Errorf("skill note = %q", here[0].Note)
	}
	if here[0].CTA != "inventory" {
		t.Errorf("skill cta = %q, want inventory", here[0].CTA)
	}
}

func TestScopeNoteOmitsZeroSide(t *testing.T) {
	if got := scopeNote(2, 0); got != "2 user" {
		t.Errorf("scopeNote(2,0) = %q, want \"2 user\"", got)
	}
	if got := scopeNote(0, 3); got != "3 project" {
		t.Errorf("scopeNote(0,3) = %q, want \"3 project\"", got)
	}
	if got := scopeNote(0, 0); got != "" {
		t.Errorf("scopeNote(0,0) = %q, want empty", got)
	}
}
