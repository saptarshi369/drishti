package claude

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/saptarshi369/drishti/internal/model"
)

// commandFrontmatter is the subset of a command file's YAML frontmatter we read.
type commandFrontmatter struct {
	Description  string   `yaml:"description"`
	Tools        []string `yaml:"allowed-tools"`
	ArgumentHint string   `yaml:"argument-hint"`
}

// ParseCommand turns one .claude/commands/*.md file into a CatCommand item. The
// command name is the file name without ".md" (Claude's own rule). Frontmatter is
// optional — an absent or malformed block yields empty attrs, never an error.
func ParseCommand(content []byte, scope model.Scope, relPath string) model.InventoryItem {
	var fm commandFrontmatter
	_ = parseFrontmatter(content, &fm)
	name := strings.TrimSuffix(filepath.Base(relPath), ".md")
	return model.InventoryItem{
		AgentCode: "claude",
		Category:  model.CatCommand,
		Name:      name,
		Scope:     scope,
		RelPath:   relPath,
		Enabled:   true,
		Attrs: map[string]string{
			"description":   fm.Description,
			"allowed_tools": strings.Join(fm.Tools, ", "),
			"argument_hint": fm.ArgumentHint,
		},
	}
}

// discoverCommands reads <base>/commands/*.md. A missing directory is not an
// error (§14). Subdirectory namespacing is deferred (see the 1b design).
func discoverCommands(base string, scope model.Scope) []model.InventoryItem {
	entries, err := os.ReadDir(filepath.Join(base, "commands"))
	if err != nil {
		return nil
	}
	var out []model.InventoryItem
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		p := filepath.Join(base, "commands", e.Name())
		b, rerr := os.ReadFile(p)
		if rerr != nil {
			continue
		}
		out = append(out, ParseCommand(b, scope, filepath.Join("commands", e.Name())))
	}
	return out
}
