package claude

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func TestParseSkill_Frontmatter(t *testing.T) {
	content := []byte("---\nname: deploy\ndescription: Ship to prod\nallowed-tools: [Bash]\n---\n# steps\n")
	got := ParseSkill(content, model.ScopeUser, "skills/deploy/SKILL.md")
	if got.Category != model.CatSkill || got.Name != "deploy" {
		t.Fatalf("category/name = %s/%s", got.Category, got.Name)
	}
	if got.Scope != model.ScopeUser || !got.Enabled {
		t.Errorf("scope/enabled = %s/%v", got.Scope, got.Enabled)
	}
	if got.Attrs["description"] != "Ship to prod" || got.Attrs["allowed_tools"] != "Bash" {
		t.Errorf("attrs = %v", got.Attrs)
	}
}

func TestParseSkill_NameFallsBackToDir(t *testing.T) {
	got := ParseSkill([]byte("# no frontmatter\n"), model.ScopeProject, "skills/changelog/SKILL.md")
	if got.Name != "changelog" {
		t.Errorf("name = %q, want changelog (from dir)", got.Name)
	}
}
