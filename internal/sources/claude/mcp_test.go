package claude

import (
	"fmt"
	"strings"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func TestParseMCP_StdioAndHTTP(t *testing.T) {
	content := []byte(`{"mcpServers":{
	  "github":{"command":"npx","args":["-y","server-github"]},
	  "stripe":{"type":"http","url":"https://mcp.stripe.com"}
	}}`)
	items, _, err := ParseMCP(content, model.ScopeUser, ".mcp.json", SecretMatcher{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	byName := map[string]model.InventoryItem{}
	for _, it := range items {
		if it.Category != model.CatMCP {
			t.Errorf("category = %s", it.Category)
		}
		byName[it.Name] = it
	}
	if byName["github"].Attrs["transport"] != "stdio" || byName["github"].Attrs["command"] != "npx" {
		t.Errorf("github attrs = %v", byName["github"].Attrs)
	}
	if byName["stripe"].Attrs["transport"] != "http" || byName["stripe"].Attrs["command"] != "https://mcp.stripe.com" {
		t.Errorf("stripe attrs = %v", byName["stripe"].Attrs)
	}
}

func TestParseMCP_Empty(t *testing.T) {
	items, _, err := ParseMCP([]byte(`{}`), model.ScopeUser, ".mcp.json", SecretMatcher{})
	if err != nil || len(items) != 0 {
		t.Fatalf("got %d items, err %v; want 0, nil", len(items), err)
	}
}

func TestParseMCP_EnvShape_ScrubsValues(t *testing.T) {
	doc := []byte(`{"mcpServers":{"db":{"command":"x","env":{"API_KEY":"sk-FAKEFAKEFAKE000","HOST":"localhost"}}}}`)
	items, shapes, err := ParseMCP(doc, model.ScopeProject, ".mcp.json", SecretMatcher{})
	if err != nil {
		t.Fatal(err)
	}
	if len(shapes) != 1 || shapes[0].Server != "db" {
		t.Fatalf("shapes = %+v, want one for db", shapes)
	}
	if len(shapes[0].SecretKeys) != 1 || shapes[0].SecretKeys[0] != "API_KEY" {
		t.Fatalf("SecretKeys = %v, want [API_KEY]", shapes[0].SecretKeys)
	}
	// Privacy: the secret value must appear nowhere in the returned data.
	blob := fmt.Sprintf("%+v %+v", items, shapes)
	if strings.Contains(blob, "sk-FAKEFAKEFAKE000") {
		t.Fatal("secret value leaked into ParseMCP output")
	}
}
