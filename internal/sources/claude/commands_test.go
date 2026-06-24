package claude

import (
	"path/filepath"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func TestParseCommand(t *testing.T) {
	content := []byte("---\ndescription: Deploy the app\nargument-hint: \"[env]\"\nallowed-tools: [Bash, Read]\n---\nDeploy $ARGUMENTS\n")
	it := ParseCommand(content, model.ScopeUser, "commands/deploy.md")
	if it.Category != model.CatCommand {
		t.Errorf("category = %q", it.Category)
	}
	if it.Name != "deploy" {
		t.Errorf("name = %q", it.Name)
	}
	if it.Attrs["description"] != "Deploy the app" {
		t.Errorf("description = %q", it.Attrs["description"])
	}
	if it.Attrs["allowed_tools"] != "Bash, Read" {
		t.Errorf("allowed_tools = %q", it.Attrs["allowed_tools"])
	}
	if it.Attrs["argument_hint"] != "[env]" {
		t.Errorf("argument_hint = %q", it.Attrs["argument_hint"])
	}
}

func TestDiscover_Commands(t *testing.T) {
	dir := t.TempDir()
	userClaude := filepath.Join(dir, "user", ".claude")
	proj := filepath.Join(dir, "proj")
	writeFile(t, filepath.Join(userClaude, "commands", "deploy.md"), "---\ndescription: d\n---\n")
	writeFile(t, filepath.Join(proj, ".claude", "commands", "lint.md"), "body\n")

	items, _, _, err := Discover(Locations{UserClaudeDir: userClaude, ProjectRoot: proj}, SecretMatcher{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	var cmds int
	for _, it := range items {
		if it.Category == model.CatCommand {
			cmds++
		}
	}
	if cmds != 2 {
		t.Errorf("command items = %d, want 2", cmds)
	}
}
