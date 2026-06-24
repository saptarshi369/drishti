package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/skills"
	"github.com/saptarshi369/drishti/internal/store"
)

// seedSkill resolves one active skill with the given tokens and fires it n times,
// so /api/skills has real data to serve.
func seedSkill(t *testing.T, st *store.Store, name string, tokens, fires int) {
	t.Helper()
	items := []model.InventoryItem{{AgentCode: "claude", Category: model.CatSkill, Name: name, Scope: model.ScopeUser, Enabled: true}}
	if err := st.ReplaceInventory("claude", "", items); err != nil {
		t.Fatal(err)
	}
	if err := st.ReplaceResolved("claude", "", []model.ResolvedItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: name, EffectiveStatus: model.StatusActive, Winner: &items[0], EstContextTokens: tokens},
	}); err != nil {
		t.Fatal(err)
	}
	evs := make([]model.Event, fires)
	for i := 0; i < fires; i++ {
		evs[i] = model.Event{AgentCode: "claude", TypeCode: "skill", SourceCode: "transcript", SessionID: "s", TsMs: int64(i + 1), SkillName: name, DedupeKey: name + string(rune('a'+i))}
	}
	if _, err := st.ApplyIngest(store.IngestBatch{Events: evs}); err != nil {
		t.Fatal(err)
	}
}

// TestHandleSkills_OK seeds a skill and verifies the handler returns a 200 with a
// snapshot reflecting it.
func TestHandleSkills_OK(t *testing.T) {
	srv := newTestServer(t)
	srv.SetSkillThresholds(skills.Thresholds{HighTriggerMin: 20, LowValueRatioMax: 5.0})
	seedSkill(t, srv.st, "deploy", 500, 3)

	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/skills", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var snap model.SkillsSnapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snap); err != nil {
		t.Fatal(err)
	}
	if snap.Total != 1 || len(snap.Items) != 1 || snap.Items[0].Name != "deploy" || snap.Items[0].Triggers != 3 {
		t.Fatalf("snapshot = %+v", snap)
	}
}
