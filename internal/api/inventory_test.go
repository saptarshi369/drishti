package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/store"
)

// TestInventoryEmptyResult verifies that an empty store returns [] (not null) in JSON.
func TestInventoryEmptyResult(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/t.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	srv := NewServer("test", st)
	req := httptest.NewRequest(http.MethodGet, "/api/inventory", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	// The body must contain "items":[] not "items":null.
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	arr, ok := m["items"].([]any)
	if !ok {
		t.Fatalf("items field is not an array; body = %s", rec.Body.String())
	}
	if len(arr) != 0 {
		t.Fatalf("expected empty array, got %d items", len(arr))
	}
}

// TestInventoryWhyBadID verifies that a non-numeric {id} returns 400.
func TestInventoryWhyBadID(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/t.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	srv := NewServer("test", st)
	req := httptest.NewRequest(http.MethodGet, "/api/inventory/notanumber/why", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d; body = %s", rec.Code, rec.Body.String())
	}
}

// TestInventoryHandler seeds the store with one resolved skill and verifies
// that GET /api/inventory?category=skill returns it in the "items" array.
func TestInventoryHandler(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/t.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	items := []model.InventoryItem{{
		AgentCode: "claude",
		Category:  model.CatSkill,
		Name:      "deploy",
		Scope:     model.ScopeUser,
		Enabled:   true,
		Attrs:     map[string]string{},
	}}
	_ = st.ReplaceInventory("claude", "", items)
	_ = st.ReplaceResolved("claude", "", []model.ResolvedItem{{
		AgentCode:       "claude",
		Category:        model.CatSkill,
		Name:            "deploy",
		EffectiveStatus: model.StatusActive,
		Winner:          &items[0],
	}})

	srv := NewServer("test", st)
	req := httptest.NewRequest(http.MethodGet, "/api/inventory?category=skill", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Items []model.ResolvedRow `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Items) != 1 || body.Items[0].Name != "deploy" {
		t.Fatalf("items = %+v", body.Items)
	}
}

// TestInventoryDefaultRoot verifies that when the request omits ?root=, the
// handler uses the server's configured default root (the daemon's pwd project
// root), which holds the merged user+project resolved set.
func TestInventoryDefaultRoot(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/t.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	items := []model.InventoryItem{{
		AgentCode: "claude", Category: model.CatMemory, Name: "CLAUDE.md (project)",
		Scope: model.ScopeProject, Enabled: true, Attrs: map[string]string{},
	}}
	_ = st.ReplaceInventory("claude", "/proj", items)
	_ = st.ReplaceResolved("claude", "/proj", []model.ResolvedItem{{
		AgentCode: "claude", Category: model.CatMemory, Name: "CLAUDE.md (project)",
		EffectiveStatus: model.StatusActive, Winner: &items[0],
	}})

	srv := NewServer("test", st)
	srv.SetDefaultRoot("/proj")
	req := httptest.NewRequest(http.MethodGet, "/api/inventory?category=memory", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	var body struct {
		Items []model.ResolvedRow `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Items) != 1 || body.Items[0].Name != "CLAUDE.md (project)" {
		t.Fatalf("default-root query returned %+v", body.Items)
	}
}
