package claude

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func TestParseAgent(t *testing.T) {
	got := ParseAgent([]byte("---\nname: reviewer\nmodel: sonnet\n---\nYou review code.\n"),
		model.ScopeUser, "agents/reviewer.md")
	if got.Category != model.CatAgent || got.Name != "reviewer" {
		t.Fatalf("category/name = %s/%s", got.Category, got.Name)
	}
	if got.Attrs["model"] != "sonnet" {
		t.Errorf("model = %q", got.Attrs["model"])
	}
}

func TestParseAgent_NameFallsBackToFile(t *testing.T) {
	got := ParseAgent([]byte("no frontmatter\n"), model.ScopeProject, "agents/explorer.md")
	if got.Name != "explorer" {
		t.Errorf("name = %q, want explorer", got.Name)
	}
}

func TestParseAgent_CapturesDescription(t *testing.T) {
	got := ParseAgent([]byte("---\nname: reviewer\nmodel: sonnet\ndescription: Reviews Go code for bugs.\n---\nbody\n"),
		model.ScopeUser, "agents/reviewer.md")
	if got.Attrs["description"] != "Reviews Go code for bugs." {
		t.Errorf("description = %q, want %q", got.Attrs["description"], "Reviews Go code for bugs.")
	}
}
