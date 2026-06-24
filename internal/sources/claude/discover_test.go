package claude

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscover_UserAndProject(t *testing.T) {
	dir := t.TempDir()
	userClaude := filepath.Join(dir, "user", ".claude")
	proj := filepath.Join(dir, "proj")

	writeFile(t, filepath.Join(userClaude, "skills", "deploy", "SKILL.md"), "---\nname: deploy\ndescription: ship\n---\n")
	writeFile(t, filepath.Join(userClaude, "agents", "reviewer.md"), "---\nname: reviewer\nmodel: sonnet\n---\n")
	writeFile(t, filepath.Join(dir, "user", ".claude.json"), `{"mcpServers":{"github":{"command":"npx"}}}`)
	writeFile(t, filepath.Join(proj, ".claude", "skills", "deploy", "SKILL.md"), "---\nname: deploy\ndescription: ship-proj\n---\n")
	writeFile(t, filepath.Join(proj, ".mcp.json"), `{"mcpServers":{"sentry":{"type":"http","url":"http://x"}}}`)
	writeFile(t, filepath.Join(proj, ".claude", "settings.json"), `{"disableBundledSkills":true}`)

	items, tg, _, err := Discover(Locations{
		UserClaudeDir:  userClaude,
		UserClaudeJSON: filepath.Join(dir, "user", ".claude.json"),
		ProjectRoot:    proj,
	}, SecretMatcher{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	count := map[model.Category]int{}
	scopes := map[string]bool{}
	for _, it := range items {
		count[it.Category]++
		scopes[string(it.Scope)] = true
	}
	if count[model.CatSkill] != 2 { // user + project deploy
		t.Errorf("skills = %d, want 2", count[model.CatSkill])
	}
	if count[model.CatMCP] != 2 {
		t.Errorf("mcp = %d, want 2", count[model.CatMCP])
	}
	if count[model.CatAgent] != 1 {
		t.Errorf("agents = %d, want 1", count[model.CatAgent])
	}
	if !scopes["user"] || !scopes["project"] {
		t.Errorf("scopes = %v", scopes)
	}
	if !tg.DisableBundledSkills {
		t.Error("project toggle not merged")
	}
}

func TestDiscover_NoProject(t *testing.T) {
	dir := t.TempDir()
	_, _, _, err := Discover(Locations{UserClaudeDir: filepath.Join(dir, "missing"), ProjectRoot: ""}, SecretMatcher{})
	if err != nil {
		t.Fatalf("missing dirs must not error: %v", err)
	}
}

func TestDiscover_Memory(t *testing.T) {
	dir := t.TempDir()
	userClaude := filepath.Join(dir, "user", ".claude")
	proj := filepath.Join(dir, "proj")

	writeFile(t, filepath.Join(userClaude, "CLAUDE.md"), "# user mem\n")
	writeFile(t, filepath.Join(userClaude, "rules", "style.md"), "use tabs\n")
	writeFile(t, filepath.Join(proj, "CLAUDE.md"), "# proj mem\n")
	writeFile(t, filepath.Join(proj, "CLAUDE.local.md"), "local\n")
	writeFile(t, filepath.Join(proj, ".claude", "CLAUDE.md"), "# dotclaude mem\n")
	writeFile(t, filepath.Join(proj, ".claude", "rules", "testing.md"), "test rule\n")

	items, _, _, err := Discover(Locations{UserClaudeDir: userClaude, ProjectRoot: proj}, SecretMatcher{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	var mem int
	for _, it := range items {
		if it.Category == model.CatMemory {
			mem++
		}
	}
	// user CLAUDE.md + user rules/style.md + proj CLAUDE.md + CLAUDE.local.md
	// + .claude/CLAUDE.md + .claude/rules/testing.md = 6
	if mem != 6 {
		t.Errorf("memory items = %d, want 6", mem)
	}
}

func TestDiscover_PluginsAndOutputStyleToggle(t *testing.T) {
	dir := t.TempDir()
	userClaude := filepath.Join(dir, "user", ".claude")
	writeFile(t, filepath.Join(userClaude, "settings.json"),
		`{"outputStyle":"Explanatory","enabledPlugins":{"github@official":true,"legacy@community":false}}`)

	items, tg, _, err := Discover(Locations{UserClaudeDir: userClaude}, SecretMatcher{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	var plugins int
	for _, it := range items {
		if it.Category == model.CatPlugin {
			plugins++
		}
	}
	if plugins != 2 {
		t.Errorf("plugin items = %d, want 2", plugins)
	}
	if tg.OutputStyle != "Explanatory" {
		t.Errorf("merged OutputStyle = %q, want Explanatory", tg.OutputStyle)
	}
}

// TestDiscover_CollectsSecurityInputs verifies that Discover aggregates
// ScopePermissions from settings files and MCPEnvShapes from MCP JSON files
// into the returned SecurityInputs.
func TestDiscover_CollectsSecurityInputs(t *testing.T) {
	dir := t.TempDir()
	// user ~/.claude/settings.json with a bypass mode + broad allow
	cdir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(cdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cdir, "settings.json"),
		[]byte(`{"permissions":{"allow":["Bash(*)"],"defaultMode":"bypassPermissions"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// ~/.claude.json with an MCP server holding a secret-shaped env value
	if err := os.WriteFile(filepath.Join(dir, ".claude.json"),
		[]byte(`{"mcpServers":{"db":{"command":"x","env":{"TOKEN":"ghp_0000000000FAKE"}}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, sec, err := Discover(Locations{UserClaudeDir: cdir, UserClaudeJSON: filepath.Join(dir, ".claude.json")}, SecretMatcher{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sec.Permissions) != 1 || sec.Permissions[0].DefaultMode != "bypassPermissions" {
		t.Fatalf("Permissions = %+v", sec.Permissions)
	}
	if len(sec.MCPEnv) != 1 || sec.MCPEnv[0].SecretKeys[0] != "TOKEN" {
		t.Fatalf("MCPEnv = %+v", sec.MCPEnv)
	}
}
