package services

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func TestBuildContextBudget(t *testing.T) {
	rows := []model.ResolvedRow{
		{ID: 1, Category: "memory", Name: "CLAUDE.md", EffectiveStatus: "active", WinnerScope: "user", EstContextTokens: 600},
		{ID: 2, Category: "mcp", Name: "github", EffectiveStatus: "active", WinnerScope: "user", EstContextTokens: 500},
		{ID: 3, Category: "skill", Name: "lint", EffectiveStatus: "active", WinnerScope: "project", EstContextTokens: 5},
		// Non-active rows must be excluded from the tax entirely.
		{ID: 4, Category: "command", Name: "deploy", EffectiveStatus: "shadowed", WinnerScope: "user", EstContextTokens: 9},
		{ID: 5, Category: "skill", Name: "old", EffectiveStatus: "disabled", WinnerScope: "user", EstContextTokens: 9},
	}
	snap := BuildContextBudget(rows, 200000)

	if snap.TotalTokens != 1105 { // 600+500+5
		t.Errorf("total = %d, want 1105", snap.TotalTokens)
	}
	if snap.WindowTokens != 200000 {
		t.Errorf("window = %d, want 200000", snap.WindowTokens)
	}
	// by_category sorted tokens desc: memory(600) > mcp(500) > skill(5).
	if len(snap.ByCategory) != 3 || snap.ByCategory[0].Category != "memory" || snap.ByCategory[2].Category != "skill" {
		t.Fatalf("by_category = %+v", snap.ByCategory)
	}
	if snap.ByCategory[0].Count != 1 || snap.ByCategory[0].Tokens != 600 {
		t.Errorf("memory bucket = %+v", snap.ByCategory[0])
	}
	// consumers: only the 3 active items, sorted tokens desc, highest first.
	if len(snap.Consumers) != 3 || snap.Consumers[0].ID != 1 || snap.Consumers[2].Name != "lint" {
		t.Fatalf("consumers = %+v", snap.Consumers)
	}
	// MCP present → an honesty caveat is emitted.
	if len(snap.Caveats) == 0 {
		t.Error("expected an MCP-estimate caveat")
	}
}

func TestBuildContextBudget_Empty(t *testing.T) {
	snap := BuildContextBudget(nil, 200000)
	if snap.TotalTokens != 0 || snap.Pct != 0 {
		t.Errorf("empty total/pct = %d/%v, want 0/0", snap.TotalTokens, snap.Pct)
	}
	// Slices must be non-nil so the API serialises [] not null.
	if snap.ByCategory == nil || snap.Consumers == nil || snap.Caveats == nil {
		t.Errorf("empty snapshot slices must be non-nil: %+v", snap)
	}
}
