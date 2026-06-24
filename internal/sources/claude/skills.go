package claude

import (
	"path/filepath"
	"strings"

	"github.com/saptarshi369/drishti/internal/model"
)

// skillFrontmatter is the subset of SKILL.md frontmatter we read.
type skillFrontmatter struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tools       []string `yaml:"allowed-tools"`
}

// ParseSkill turns one SKILL.md into an InventoryItem. The name comes from the
// frontmatter when present, otherwise from the skill's directory name (Claude's
// own fallback). relPath is relative to the scope root, for display + identity.
func ParseSkill(content []byte, scope model.Scope, relPath string) model.InventoryItem {
	var fm skillFrontmatter
	_ = parseFrontmatter(content, &fm) // absent frontmatter is fine; we fall back
	name := fm.Name
	if name == "" {
		name = filepath.Base(filepath.Dir(relPath))
	}
	return model.InventoryItem{
		AgentCode: "claude",
		Category:  model.CatSkill,
		Name:      name,
		Scope:     scope,
		RelPath:   relPath,
		Enabled:   true,
		Attrs: map[string]string{
			"description":   fm.Description,
			"allowed_tools": strings.Join(fm.Tools, ", "),
		},
	}
}
