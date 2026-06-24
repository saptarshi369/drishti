package store

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	st, err := Open(t.TempDir() + "/t.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func TestReplaceInventoryAndResolved(t *testing.T) {
	st := tempStore(t)
	items := []model.InventoryItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "deploy", Scope: model.ScopeUser, RelPath: "skills/deploy/SKILL.md", Enabled: true, Attrs: map[string]string{"description": "ship"}},
	}
	if err := st.ReplaceInventory("claude", "", items); err != nil {
		t.Fatalf("ReplaceInventory: %v", err)
	}
	resolved := []model.ResolvedItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "deploy", EffectiveStatus: model.StatusActive,
			Winner: &items[0], EstContextTokens: 2,
			PrecedenceTrail: []model.PrecedenceStep{{Step: 1, Scope: "user", Decision: "wins", Reason: "user scope wins"}}},
	}
	if err := st.ReplaceResolved("claude", "", resolved); err != nil {
		t.Fatalf("ReplaceResolved: %v", err)
	}

	// Re-running must be idempotent (replace, not duplicate).
	if err := st.ReplaceInventory("claude", "", items); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := st.DB().QueryRow(`SELECT count(*) FROM inventory_items`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("inventory_items rows = %d, want 1", n)
	}
	var rn int
	if err := st.DB().QueryRow(`SELECT count(*) FROM inventory_resolved`).Scan(&rn); err != nil {
		t.Fatal(err)
	}
	if rn != 0 {
		t.Errorf("inventory_resolved rows after re-run = %d, want 0 (replaced items wipe derived resolved)", rn)
	}
}

func TestListResolvedAndTrail(t *testing.T) {
	st := tempStore(t)
	items := []model.InventoryItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "deploy", Scope: model.ScopeUser, Enabled: true, Attrs: map[string]string{"description": "ship"}},
		{AgentCode: "claude", Category: model.CatSkill, Name: "pdf", Scope: model.ScopeUser, Enabled: true, Attrs: map[string]string{}},
	}
	_ = st.ReplaceInventory("claude", "", items)
	resolved := []model.ResolvedItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "deploy", EffectiveStatus: model.StatusActive, Winner: &items[0],
			PrecedenceTrail: []model.PrecedenceStep{{Step: 1, Scope: "user", Decision: "wins", Reason: "user wins"}}},
		{AgentCode: "claude", Category: model.CatSkill, Name: "pdf", EffectiveStatus: model.StatusDisabled},
	}
	_ = st.ReplaceResolved("claude", "", resolved)

	active, err := st.ListResolved("skill", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || active[0].Name != "deploy" {
		t.Fatalf("active rows = %+v, want only deploy", active)
	}
	if !active[0].InUser || active[0].WinnerScope != "user" {
		t.Errorf("row = %+v", active[0])
	}
	all, err := st.ListResolved("skill", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("showDisabled rows = %d, want 2", len(all))
	}

	steps, err := st.ResolvedTrail(active[0].ID)
	if err != nil || len(steps) != 1 || steps[0].Decision != "wins" {
		t.Errorf("trail = %+v err %v", steps, err)
	}
}

func TestResolvedTrailUnknownID(t *testing.T) {
	st := tempStore(t)
	// Query for a non-existent row; should return nil, nil (not an error).
	steps, err := st.ResolvedTrail(99999)
	if err != nil || steps != nil {
		t.Errorf("ResolvedTrail for unknown ID = %+v, %v; want nil, nil", steps, err)
	}
}

func TestResolvedTrailMalformedJSON(t *testing.T) {
	st := tempStore(t)
	items := []model.InventoryItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "trail-test", Scope: model.ScopeUser, Enabled: true},
	}
	_ = st.ReplaceInventory("claude", "", items)
	resolved := []model.ResolvedItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "trail-test", EffectiveStatus: model.StatusActive, Winner: &items[0],
			PrecedenceTrail: []model.PrecedenceStep{{Step: 1, Scope: "user", Decision: "wins"}}},
	}
	_ = st.ReplaceResolved("claude", "", resolved)

	var rowID int64
	_ = st.db.QueryRow(`SELECT id FROM inventory_resolved WHERE name=?`, "trail-test").Scan(&rowID)

	// Corrupt the precedence_trail to invalid JSON.
	_, _ = st.db.Exec(`UPDATE inventory_resolved SET precedence_trail='[bad json' WHERE id=?`, rowID)

	_, err := st.ResolvedTrail(rowID)
	if err == nil {
		t.Fatal("ResolvedTrail should error on malformed JSON; got nil")
	}
}

func TestListResolvedMalformedAttrs(t *testing.T) {
	st := tempStore(t)
	items := []model.InventoryItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "bad", Scope: model.ScopeUser, Enabled: true, Attrs: map[string]string{"x": "y"}},
	}
	_ = st.ReplaceInventory("claude", "", items)
	resolved := []model.ResolvedItem{
		{AgentCode: "claude", Category: model.CatSkill, Name: "bad", EffectiveStatus: model.StatusActive, Winner: &items[0]},
	}
	_ = st.ReplaceResolved("claude", "", resolved)

	// Corrupt the attrs in the inventory_items table to invalid JSON (for the winner).
	_, _ = st.db.Exec(`UPDATE inventory_items SET attrs='{bad:1' WHERE name=? AND scope=?`, "bad", "user")

	_, err := st.ListResolved("skill", "", false)
	if err == nil {
		t.Fatal("ListResolved should error on malformed winner_attrs; got nil")
	}
}
