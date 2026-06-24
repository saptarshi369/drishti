package claude

import (
	"encoding/json"
	"sort"

	"github.com/saptarshi369/drishti/internal/model"
)

// mcpFile is the subset of .mcp.json / ~/.claude.json we read.
type mcpFile struct {
	MCPServers map[string]mcpServer `json:"mcpServers"`
}

// mcpServer holds the fields we read from one entry in the mcpServers map.
// Env is decoded here so we can inspect key names for secret detection, but
// the VALUES are never stored beyond that classification step.
type mcpServer struct {
	Command string            `json:"command"`
	Type    string            `json:"type"`
	URL     string            `json:"url"`
	Env     map[string]string `json:"env"`
}

// ParseMCP reads the mcpServers object into one InventoryItem per server and,
// separately, the privacy-safe SHAPE of each server's env: the NAMES of keys
// whose values look like secrets (per sm). Env VALUES are inspected only to make
// that decision and are never returned, stored, or logged (privacy default D8).
// Servers and their secret-key lists are returned in deterministic sorted order.
// A malformed document returns an error to the caller.
func ParseMCP(content []byte, scope model.Scope, relPath string, sm SecretMatcher) ([]model.InventoryItem, []model.MCPEnvShape, error) {
	var f mcpFile
	if err := json.Unmarshal(content, &f); err != nil {
		return nil, nil, err
	}

	// Sort server names so output is deterministic for tests and database inserts.
	names := make([]string, 0, len(f.MCPServers))
	for name := range f.MCPServers {
		names = append(names, name)
	}
	sort.Strings(names)

	items := make([]model.InventoryItem, 0, len(names))
	var shapes []model.MCPEnvShape

	for _, name := range names {
		srv := f.MCPServers[name]

		// Determine transport type and endpoint address.
		// stdio servers have a command; http/sse servers have a URL.
		transport, endpoint := "stdio", srv.Command
		if srv.URL != "" {
			endpoint = srv.URL
			transport = srv.Type
			if transport == "" {
				transport = "http"
			}
		}
		items = append(items, model.InventoryItem{
			AgentCode: "claude",
			Category:  model.CatMCP,
			Name:      name,
			Scope:     scope,
			RelPath:   relPath,
			Enabled:   true,
			Attrs:     map[string]string{"transport": transport, "command": endpoint},
		})

		// Env shape: iterate env keys in sorted order so SecretKeys is
		// deterministic. Pass each (key, value) to sm.Match to decide if the
		// VALUE looks like a secret; if yes, record the KEY NAME only.
		// The value is used only for classification and is never stored or returned.
		envKeys := make([]string, 0, len(srv.Env))
		for k := range srv.Env {
			envKeys = append(envKeys, k)
		}
		sort.Strings(envKeys)

		var secretKeys []string
		for _, k := range envKeys {
			if sm.Match(k, srv.Env[k]) {
				// Only the KEY NAME is kept; the value is deliberately discarded.
				secretKeys = append(secretKeys, k)
			}
		}
		// Only emit a shape entry when at least one secret-like key exists.
		if len(secretKeys) > 0 {
			shapes = append(shapes, model.MCPEnvShape{
				Server:     name,
				Scope:      scope,
				RelPath:    relPath,
				SecretKeys: secretKeys,
			})
		}
	}
	return items, shapes, nil
}
