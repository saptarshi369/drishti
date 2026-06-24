package claude

import (
	"os"
	"path/filepath"

	"github.com/saptarshi369/drishti/internal/model"
)

// Locations tells Discover where each scope's files live. Empty fields are
// skipped. Kept explicit (not a single root) because Claude scatters config:
// MCP user/local servers live in ~/.claude.json, not under ~/.claude.
type Locations struct {
	UserClaudeDir  string // e.g. ~/.claude
	UserClaudeJSON string // e.g. ~/.claude.json (MCP user/local scope)
	ProjectRoot    string // e.g. <root>; "" disables project scope
}

// Discover walks the user and project trees and returns every raw InventoryItem,
// the merged Toggles (local > project > user), and a SecurityInputs snapshot
// containing permissions, MCP env shapes, and plugin sources. All filesystem
// errors for missing paths are swallowed: an absent file means "no items here",
// never a failure (§14). Project items carry ProjectRoot so resolution is
// per-project. sm is threaded through to the MCP and settings parsers so they
// can classify secret-shaped env values.
func Discover(loc Locations, sm SecretMatcher) ([]model.InventoryItem, model.Toggles, model.SecurityInputs, error) {
	var items []model.InventoryItem
	var userTg, projTg, localTg model.Toggles
	// sec accumulates the security-relevant data parsed from each config file.
	// Its sub-slices start nil and grow only when a file is actually present.
	var sec model.SecurityInputs

	if loc.UserClaudeDir != "" {
		items = append(items, discoverSkills(loc.UserClaudeDir, model.ScopeUser)...)
		items = append(items, discoverAgents(loc.UserClaudeDir, model.ScopeUser)...)
		items = append(items, discoverMemoryFile(filepath.Join(loc.UserClaudeDir, "CLAUDE.md"), "CLAUDE.md", model.ScopeUser)...)
		items = append(items, discoverMemoryRules(filepath.Join(loc.UserClaudeDir, "rules"), model.ScopeUser)...)
		items = append(items, discoverCommands(loc.UserClaudeDir, model.ScopeUser)...)
		items = append(items, discoverOutputStyles(loc.UserClaudeDir, model.ScopeUser)...)
		// discoverSettings now returns ScopePermissions and a presence flag (ok).
		// We only append permissions when ok=true so absent files are not represented.
		h, tg, perms, ok := discoverSettings(filepath.Join(loc.UserClaudeDir, "settings.json"), model.ScopeUser, sm)
		items = append(items, h...)
		userTg = tg
		if ok {
			sec.Permissions = append(sec.Permissions, perms)
		}
	}
	if loc.UserClaudeJSON != "" {
		// discoverMCP now returns MCPEnvShapes alongside the inventory items.
		mcpItems, shapes := discoverMCP(loc.UserClaudeJSON, model.ScopeUser, sm)
		items = append(items, mcpItems...)
		sec.MCPEnv = append(sec.MCPEnv, shapes...)
	}
	if root := loc.ProjectRoot; root != "" {
		pc := filepath.Join(root, ".claude")
		items = append(items, withRoot(discoverSkills(pc, model.ScopeProject), root)...)
		items = append(items, withRoot(discoverAgents(pc, model.ScopeProject), root)...)
		items = append(items, withRoot(discoverMemoryFile(filepath.Join(root, "CLAUDE.md"), "CLAUDE.md", model.ScopeProject), root)...)
		items = append(items, withRoot(discoverMemoryFile(filepath.Join(root, "CLAUDE.local.md"), "CLAUDE.local.md", model.ScopeLocal), root)...)
		items = append(items, withRoot(discoverMemoryFile(filepath.Join(pc, "CLAUDE.md"), ".claude/CLAUDE.md", model.ScopeProject), root)...)
		items = append(items, withRoot(discoverMemoryRules(filepath.Join(pc, "rules"), model.ScopeProject), root)...)
		items = append(items, withRoot(discoverCommands(pc, model.ScopeProject), root)...)
		items = append(items, withRoot(discoverOutputStyles(pc, model.ScopeProject), root)...)

		mcpItems, shapes := discoverMCP(filepath.Join(root, ".mcp.json"), model.ScopeProject, sm)
		items = append(items, withRoot(mcpItems, root)...)
		sec.MCPEnv = append(sec.MCPEnv, shapes...)

		ph, ptg, pperms, pok := discoverSettings(filepath.Join(pc, "settings.json"), model.ScopeProject, sm)
		items = append(items, withRoot(ph, root)...)
		projTg = ptg
		if pok {
			sec.Permissions = append(sec.Permissions, pperms)
		}
		lh, ltg, lperms, lok := discoverSettings(filepath.Join(pc, "settings.local.json"), model.ScopeLocal, sm)
		items = append(items, withRoot(lh, root)...)
		localTg = ltg
		if lok {
			sec.Permissions = append(sec.Permissions, lperms)
		}
	}

	// Plugin sources for the untrusted-source rule are derived from the plugin
	// inventory items already discovered (they carry the "marketplace" attr set
	// by ParseSettings). Only enabled plugins are flagged; disabled ones are not
	// threat-relevant.
	for _, it := range items {
		if it.Category == model.CatPlugin && it.Enabled {
			sec.Plugins = append(sec.Plugins, model.PluginSource{
				Name:        it.Name,
				Marketplace: it.Attrs["marketplace"],
				Scope:       it.Scope,
				RelPath:     it.RelPath,
			})
		}
	}

	return items, MergeToggles(localTg, projTg, userTg), sec, nil
}

// withRoot stamps ProjectRoot onto items discovered under a project tree.
func withRoot(items []model.InventoryItem, root string) []model.InventoryItem {
	for i := range items {
		items[i].ProjectRoot = root
	}
	return items
}

// discoverSkills reads <base>/skills/*/SKILL.md.
func discoverSkills(base string, scope model.Scope) []model.InventoryItem {
	var out []model.InventoryItem
	entries, err := os.ReadDir(filepath.Join(base, "skills"))
	if err != nil {
		// Missing directory is not an error per §14 (degrade, don't die).
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p := filepath.Join(base, "skills", e.Name(), "SKILL.md")
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		out = append(out, ParseSkill(b, scope, filepath.Join("skills", e.Name(), "SKILL.md")))
	}
	return out
}

// discoverAgents reads <base>/agents/**/*.md recursively.
func discoverAgents(base string, scope model.Scope) []model.InventoryItem {
	var out []model.InventoryItem
	root := filepath.Join(base, "agents")
	// WalkDir calls the callback with a non-nil err when the agents dir is missing;
	// the callback returns nil on any error, so a missing dir is silently skipped (§14).
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(p) != ".md" {
			return nil
		}
		b, rerr := os.ReadFile(p)
		if rerr != nil {
			return nil
		}
		rel, _ := filepath.Rel(base, p)
		out = append(out, ParseAgent(b, scope, rel))
		return nil
	})
	return out
}

// discoverMCP reads one MCP JSON file and returns both the inventory items and
// the MCPEnvShapes (one per server whose env keys look secret-shaped). A
// missing file is silently skipped per §14 — returning nil, nil rather than
// an error.
func discoverMCP(path string, scope model.Scope, sm SecretMatcher) ([]model.InventoryItem, []model.MCPEnvShape) {
	b, err := os.ReadFile(path)
	if err != nil {
		// Missing file is not an error (§14).
		return nil, nil
	}
	items, shapes, err := ParseMCP(b, scope, filepath.Base(path), sm)
	if err != nil {
		return nil, nil
	}
	return items, shapes
}

// discoverSettings reads one settings file and returns the hook items, toggle
// values, permissions snapshot, and a presence flag. The presence flag (ok)
// is false when the file does not exist, so the caller can distinguish "file
// absent" (no permissions entry) from "file present but empty permissions"
// (permissions entry with zero values). Missing file is silently skipped per §14.
func discoverSettings(path string, scope model.Scope, sm SecretMatcher) ([]model.InventoryItem, model.Toggles, model.ScopePermissions, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		// Missing file is not an error (§14).
		return nil, model.Toggles{}, model.ScopePermissions{}, false
	}
	hooks, tg, perms, err := ParseSettings(b, scope, filepath.Base(path), sm)
	if err != nil {
		return nil, model.Toggles{}, model.ScopePermissions{}, false
	}
	return hooks, tg, perms, true
}
