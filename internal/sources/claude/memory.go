package claude

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/saptarshi369/drishti/internal/model"
)

// ParseMemory builds one CatMemory InventoryItem from a memory file (CLAUDE.md,
// CLAUDE.local.md, or a rules/*.md file). These are plain markdown with no
// resolution frontmatter, so we record only what the UI and Context-Budget
// module need: the byte size (drives the chars/4 token estimate), a one-line
// summary for the drawer, and the absolute path (matched against claudeMdExcludes
// at resolve time). The display Name is scope-qualified because user and project
// commonly both define "CLAUDE.md" — qualifying keeps the resolved key unique.
func ParseMemory(content []byte, scope model.Scope, relPath, absPath string) model.InventoryItem {
	return model.InventoryItem{
		AgentCode: "claude",
		Category:  model.CatMemory,
		Name:      relPath + " (" + string(scope) + ")",
		Scope:     scope,
		RelPath:   relPath,
		Enabled:   true,
		Attrs: map[string]string{
			"bytes":   strconv.Itoa(len(content)),
			"summary": memorySummary(content),
			"abs":     absPath,
		},
	}
}

// memorySummary returns the first non-empty line with leading markdown heading
// markers ("# ") trimmed, capped at 80 chars, for a compact drawer preview.
// HTML-comment lines are skipped (Claude strips them before loading memory).
func memorySummary(content []byte) string {
	for _, line := range strings.Split(string(content), "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "<!--") {
			continue
		}
		t = strings.TrimLeft(t, "# ")
		if len(t) > 80 {
			t = t[:80]
		}
		return t
	}
	return ""
}

// discoverMemoryFile reads a single memory file if present. A missing file is not
// an error (§14): it just yields no item.
func discoverMemoryFile(absPath, relPath string, scope model.Scope) []model.InventoryItem {
	b, err := os.ReadFile(absPath)
	if err != nil {
		return nil
	}
	return []model.InventoryItem{ParseMemory(b, scope, relPath, absPath)}
}

// discoverMemoryRules reads every *.md directly under a rules/ directory. A
// missing directory is not an error (§14).
func discoverMemoryRules(dir string, scope model.Scope) []model.InventoryItem {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []model.InventoryItem
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		abs := filepath.Join(dir, e.Name())
		b, rerr := os.ReadFile(abs)
		if rerr != nil {
			continue
		}
		out = append(out, ParseMemory(b, scope, filepath.Join("rules", e.Name()), abs))
	}
	return out
}
