package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/security"
	claude "github.com/saptarshi369/drishti/internal/sources/claude"
	"github.com/saptarshi369/drishti/internal/store"
)

// fakeStore is a minimal in-memory test double for InventoryStore. It records
// every write call so tests can assert on what was persisted without a real DB.
// In Go, this pattern — implement just the interface you need — is preferred over
// mocking frameworks because it keeps tests simple and the interface contract clear.
type fakeStore struct {
	items    []model.InventoryItem
	resolved []model.ResolvedItem
	findings []model.Finding
}

// ReplaceInventory records the items written (mirrors the real store method signature).
func (f *fakeStore) ReplaceInventory(_ string, _ string, items []model.InventoryItem) error {
	f.items = append(f.items, items...)
	return nil
}

// ReplaceResolved records the resolved items written.
func (f *fakeStore) ReplaceResolved(_ string, _ string, resolved []model.ResolvedItem) error {
	f.resolved = append(f.resolved, resolved...)
	return nil
}

// ReplaceSecurityFindings records the findings written so tests can assert on them.
func (f *fakeStore) ReplaceSecurityFindings(_ string, _ string, findings []model.Finding) error {
	f.findings = append(f.findings, findings...)
	return nil
}

// TestRefreshInventory_EndToEnd exercises the full pipeline against a real store
// and a synthesised user-global location that has one skill SKILL.md.
func TestRefreshInventory_EndToEnd(t *testing.T) {
	dir := t.TempDir()
	userClaude := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(filepath.Join(userClaude, "skills", "deploy"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(userClaude, "skills", "deploy", "SKILL.md"),
		[]byte("---\nname: deploy\ndescription: ship\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	st, err := store.Open(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	locs := []claude.Locations{{UserClaudeDir: userClaude}}
	// Pass security.DefaultRules() — Task 11 will replace this with file-based load.
	if err := RefreshInventory(st, locs, security.DefaultRules()); err != nil {
		t.Fatalf("RefreshInventory: %v", err)
	}
	rows, err := st.ListResolved("skill", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Name != "deploy" || rows[0].EffectiveStatus != "active" {
		t.Fatalf("rows = %+v", rows)
	}
}

// TestRefreshInventory_PersistsFindings verifies that RefreshInventory calls the
// security rule engine on the discovered inputs and writes findings to the store.
// We use a settings.json with "bypassPermissions" which should trigger the built-in
// bypass-permissions-mode rule — giving us a reliable assertion without injecting
// synthetic inputs.
func TestRefreshInventory_PersistsFindings(t *testing.T) {
	dir := t.TempDir()
	cdir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(cdir, 0o755)
	_ = os.WriteFile(filepath.Join(cdir, "settings.json"),
		[]byte(`{"permissions":{"defaultMode":"bypassPermissions"}}`), 0o644)

	fs := &fakeStore{} // records ReplaceInventory/ReplaceResolved/ReplaceSecurityFindings
	rules := security.DefaultRules()
	locs := []claude.Locations{{UserClaudeDir: cdir}}
	if err := RefreshInventory(fs, locs, rules); err != nil {
		t.Fatal(err)
	}
	if len(fs.findings) == 0 {
		t.Fatal("expected at least one finding (bypass-permissions-mode) to be persisted")
	}
	found := false
	for _, f := range fs.findings {
		if f.RuleID == "bypass-permissions-mode" {
			found = true
		}
	}
	if !found {
		t.Fatalf("bypass finding not persisted; got %+v", fs.findings)
	}
}
