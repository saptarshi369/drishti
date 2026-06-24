package claude

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func TestParseMemory(t *testing.T) {
	content := []byte("<!-- note -->\n\n# Project rules\n\nUse tabs.\n")
	it := ParseMemory(content, model.ScopeProject, "CLAUDE.md", "/repo/CLAUDE.md")
	if it.Category != model.CatMemory {
		t.Errorf("category = %q", it.Category)
	}
	if it.Name != "CLAUDE.md (project)" {
		t.Errorf("name = %q", it.Name)
	}
	if it.Attrs["abs"] != "/repo/CLAUDE.md" {
		t.Errorf("abs = %q", it.Attrs["abs"])
	}
	if it.Attrs["summary"] != "Project rules" {
		t.Errorf("summary = %q", it.Attrs["summary"])
	}
	if it.Attrs["bytes"] == "" {
		t.Error("bytes attr missing")
	}
}
