package resolve

import (
	"strings"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

// hook builds a minimal hook InventoryItem for testing. The "command" attr
// mirrors what the real parser emits for a hook entry.
func hook(name, cmd string, scope model.Scope) model.InventoryItem {
	return model.InventoryItem{AgentCode: "claude", Category: model.CatHook, Name: name, Scope: scope, Enabled: true, Attrs: map[string]string{"command": cmd}}
}

// TestResolveHooks_AllActiveAndMerge verifies that hooks from different scopes
// all become active resolved rows (hooks MERGE — no single scope wins).
func TestResolveHooks_AllActiveAndMerge(t *testing.T) {
	items := []model.InventoryItem{
		hook("PreToolUse · Bash", "guard.sh", model.ScopeUser),
		hook("PostToolUse · Edit", "fmt.sh", model.ScopeProject),
	}
	// Exercise the hooks resolver directly: Resolve() also emits other categories
	// (e.g. built-in output styles), which would mask the hook-specific assertion.
	got := resolveHooks(items)
	if len(got) != 2 {
		t.Fatalf("got %d, want 2", len(got))
	}
	for _, r := range got {
		if r.EffectiveStatus != model.StatusActive {
			t.Errorf("%s status = %s, want active", r.Name, r.EffectiveStatus)
		}
	}
}

// TestResolveHooks_DuplicateNamesDisambiguated verifies that two hooks with the
// same display name (from different scopes) produce two unique resolved names,
// with the second suffixed " (2)" so the (category, name) key stays unique.
func TestResolveHooks_DuplicateNamesDisambiguated(t *testing.T) {
	items := []model.InventoryItem{
		hook("PreToolUse · Bash", "a.sh", model.ScopeUser),
		hook("PreToolUse · Bash", "b.sh", model.ScopeProject),
	}
	got := resolveHooks(items)
	names := map[string]bool{}
	for _, r := range got {
		if names[r.Name] {
			t.Fatalf("duplicate resolved name %q", r.Name)
		}
		names[r.Name] = true
	}
	if len(got) != 2 {
		t.Fatalf("got %d, want 2", len(got))
	}
	var suffixed bool
	for n := range names {
		if strings.HasSuffix(n, "(2)") {
			suffixed = true
		}
	}
	if !suffixed {
		t.Errorf("expected a disambiguated '(2)' name, got %v", names)
	}
}
