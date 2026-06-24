package resolve

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

// mcp is a test helper that creates a CatMCP InventoryItem for the given name
// and scope. The enabled flag is true and a minimal Attrs map is provided so
// the item looks like a real MCP server entry.
func mcp(name string, scope model.Scope) model.InventoryItem {
	return model.InventoryItem{AgentCode: "claude", Category: model.CatMCP, Name: name, Scope: scope, Enabled: true, Attrs: map[string]string{"transport": "stdio"}}
}

// TestResolveMCP_ProjectBeatsUser checks that MCP servers follow project>user
// precedence (opposite of skills, matching Claude Code's documented behavior).
func TestResolveMCP_ProjectBeatsUser(t *testing.T) {
	items := []model.InventoryItem{mcp("github", model.ScopeUser), mcp("github", model.ScopeProject)}
	got := Resolve(items, model.Toggles{})
	r, _ := find(got, "github")
	if r.EffectiveStatus != model.StatusActive || r.Winner.Scope != model.ScopeProject {
		t.Errorf("status/winner = %s/%s, want active/project", r.EffectiveStatus, r.Winner.Scope)
	}
}

// TestResolveMCP_Disabled checks that a name listed in DisabledMcpjsonServers
// resolves to StatusDisabled regardless of which scope it appears in.
func TestResolveMCP_Disabled(t *testing.T) {
	items := []model.InventoryItem{mcp("postgres", model.ScopeUser)}
	got := Resolve(items, model.Toggles{DisabledMcpjsonServers: []string{"postgres"}})
	r, _ := find(got, "postgres")
	if r.EffectiveStatus != model.StatusDisabled {
		t.Errorf("status = %s, want disabled", r.EffectiveStatus)
	}
}

// TestResolveMCP_Allowlist checks that a non-nil EnabledMcpjsonServers list
// acts as an allowlist: servers absent from the list resolve disabled, those
// present resolve active.
func TestResolveMCP_Allowlist(t *testing.T) {
	items := []model.InventoryItem{mcp("github", model.ScopeUser), mcp("sentry", model.ScopeUser)}
	got := Resolve(items, model.Toggles{EnabledMcpjsonServers: []string{"github"}})
	if r, _ := find(got, "sentry"); r.EffectiveStatus != model.StatusDisabled {
		t.Errorf("sentry status = %s, want disabled (not in allowlist)", r.EffectiveStatus)
	}
	if r, _ := find(got, "github"); r.EffectiveStatus != model.StatusActive {
		t.Errorf("github status = %s, want active", r.EffectiveStatus)
	}
}
