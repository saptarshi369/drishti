package claude

import (
	"path/filepath"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func TestParseOutputStyle(t *testing.T) {
	content := []byte("---\nname: Diagrams first\ndescription: Lead with a diagram\nkeep-coding-instructions: true\n---\nbody\n")
	it := ParseOutputStyle(content, model.ScopeUser, "output-styles/diagrams.md")
	if it.Category != model.CatOutputStyle {
		t.Errorf("category = %q", it.Category)
	}
	if it.Name != "Diagrams first" {
		t.Errorf("name = %q", it.Name)
	}
	if it.Attrs["description"] != "Lead with a diagram" {
		t.Errorf("description = %q", it.Attrs["description"])
	}
	if it.Attrs["keep_coding_instructions"] != "true" {
		t.Errorf("keep_coding = %q", it.Attrs["keep_coding_instructions"])
	}
}

func TestDiscover_OutputStyles(t *testing.T) {
	dir := t.TempDir()
	userClaude := filepath.Join(dir, "user", ".claude")
	proj := filepath.Join(dir, "proj")
	writeFile(t, filepath.Join(userClaude, "output-styles", "terse.md"), "---\nname: Terse\n---\n")
	writeFile(t, filepath.Join(proj, ".claude", "output-styles", "diagrams.md"), "body\n")

	items, _, _, err := Discover(Locations{UserClaudeDir: userClaude, ProjectRoot: proj}, SecretMatcher{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	var n int
	for _, it := range items {
		if it.Category == model.CatOutputStyle {
			n++
		}
	}
	if n != 2 {
		t.Errorf("output-style items = %d, want 2", n)
	}
}
