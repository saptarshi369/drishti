package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/store"
)

func TestHandleContextBudget(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/t.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	// Seed one active memory row so the snapshot is non-empty.
	if err := st.ReplaceInventory("claude", "", []model.InventoryItem{
		{Category: model.CatMemory, Name: "CLAUDE.md", Scope: model.ScopeUser,
			Attrs: map[string]string{"bytes": "400", "abs": "/u/CLAUDE.md"}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.ReplaceResolved("claude", "", []model.ResolvedItem{
		{Category: model.CatMemory, Name: "CLAUDE.md", EffectiveStatus: model.StatusActive,
			Winner: &model.InventoryItem{Category: model.CatMemory, Name: "CLAUDE.md", Scope: model.ScopeUser,
				Attrs: map[string]string{"bytes": "400"}}, EstContextTokens: 100},
	}); err != nil {
		t.Fatal(err)
	}

	s := NewServer("test", st)
	s.SetContextWindowTokens(200000)

	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/context-budget", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var snap model.ContextBudgetSnapshot
	if err := json.Unmarshal(rr.Body.Bytes(), &snap); err != nil {
		t.Fatal(err)
	}
	if snap.TotalTokens != 100 || snap.WindowTokens != 200000 {
		t.Errorf("total/window = %d/%d, want 100/200000", snap.TotalTokens, snap.WindowTokens)
	}
}

func TestHandleContextBudget_UnknownAgent(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/t.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	s := NewServer("test", st)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/context-budget?agent=codex", nil))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}
