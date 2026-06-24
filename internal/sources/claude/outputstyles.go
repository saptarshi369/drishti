package claude

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/saptarshi369/drishti/internal/model"
)

// outputStyleFrontmatter is the subset of an output-style file's frontmatter.
type outputStyleFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	KeepCoding  bool   `yaml:"keep-coding-instructions"`
}

// ParseOutputStyle turns one output-styles/*.md file into a CatOutputStyle item.
// The style name is the frontmatter "name", falling back to the file name.
func ParseOutputStyle(content []byte, scope model.Scope, relPath string) model.InventoryItem {
	var fm outputStyleFrontmatter
	_ = parseFrontmatter(content, &fm)
	name := fm.Name
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(relPath), ".md")
	}
	return model.InventoryItem{
		AgentCode: "claude",
		Category:  model.CatOutputStyle,
		Name:      name,
		Scope:     scope,
		RelPath:   relPath,
		Enabled:   true,
		Attrs: map[string]string{
			"description":              fm.Description,
			"keep_coding_instructions": strconv.FormatBool(fm.KeepCoding),
		},
	}
}

// discoverOutputStyles reads <base>/output-styles/*.md. Missing dir → no items.
func discoverOutputStyles(base string, scope model.Scope) []model.InventoryItem {
	entries, err := os.ReadDir(filepath.Join(base, "output-styles"))
	if err != nil {
		return nil
	}
	var out []model.InventoryItem
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		p := filepath.Join(base, "output-styles", e.Name())
		b, rerr := os.ReadFile(p)
		if rerr != nil {
			continue
		}
		out = append(out, ParseOutputStyle(b, scope, filepath.Join("output-styles", e.Name())))
	}
	return out
}
