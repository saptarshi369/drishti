package claude

import (
	"path/filepath"
	"strings"

	"github.com/saptarshi369/drishti/internal/model"
)

// agentFrontmatter is the subset of an agent .md frontmatter we read.
type agentFrontmatter struct {
	Name        string `yaml:"name"`
	Model       string `yaml:"model"`
	Description string `yaml:"description"`
}

// ParseAgent turns one agent Markdown file into an InventoryItem. Identity comes
// from the frontmatter name (Claude identifies agents by name, not path),
// falling back to the filename without its extension.
func ParseAgent(content []byte, scope model.Scope, relPath string) model.InventoryItem {
	var fm agentFrontmatter
	_ = parseFrontmatter(content, &fm)
	name := fm.Name
	if name == "" {
		base := filepath.Base(relPath)
		name = strings.TrimSuffix(base, filepath.Ext(base))
	}
	return model.InventoryItem{
		AgentCode: "claude",
		Category:  model.CatAgent,
		Name:      name,
		Scope:     scope,
		RelPath:   relPath,
		Enabled:   true,
		Attrs:     map[string]string{"model": fm.Model, "description": fm.Description},
	}
}
