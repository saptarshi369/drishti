package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestProposePreservesExistingKeys verifies that ProposeSettingsJSON preserves
// the user's existing keys and adds only the statusLine entry, returning a
// non-empty added diff.
func TestProposePreservesExistingKeys(t *testing.T) {
	existing := []byte(`{"model":"opus","env":{"FOO":"bar"}}`)
	proposed, added, err := ProposeSettingsJSON(existing, "/home/me/.drishti/bin")
	if err != nil {
		t.Fatalf("propose: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(proposed, &m); err != nil {
		t.Fatalf("proposed is not valid JSON: %v", err)
	}
	if m["model"] != "opus" {
		t.Error("existing key 'model' must be preserved (non-breaking)")
	}
	if _, ok := m["statusLine"]; !ok {
		t.Error("statusLine should be added")
	}
	if len(added) == 0 {
		t.Error("added diff should be non-empty")
	}
}

// TestProposeIsIdempotent verifies that re-proposing an already-proposed doc
// adds nothing (idempotent operation).
func TestProposeIsIdempotent(t *testing.T) {
	first, _, _ := ProposeSettingsJSON([]byte(`{}`), "/b")
	second, added, _ := ProposeSettingsJSON(first, "/b")
	if len(added) != 0 {
		t.Errorf("re-proposing an already-installed file should add nothing, got %v", added)
	}
	_ = second
}

// TestProposeRejectsInvalidJSON verifies that a malformed existing JSON returns
// an error rather than silently producing a broken proposal.
func TestProposeRejectsInvalidJSON(t *testing.T) {
	if _, _, err := ProposeSettingsJSON([]byte("{not json"), "/b"); err == nil {
		t.Error("invalid existing JSON must error")
	}
}

// TestProposeEmptyStartsFresh verifies that a nil existing slice is treated as
// {} with no error — the "fresh install" path.
func TestProposeEmptyStartsFresh(t *testing.T) {
	if _, _, err := ProposeSettingsJSON(nil, "/b"); err != nil {
		t.Errorf("nil existing should start from {} without error: %v", err)
	}
}

// TestEnsureHelperScripts verifies that EnsureHelperScripts writes the
// statusline-helper.sh script into binDir with 0o755 permissions, and is
// idempotent (calling it twice does not error).
func TestEnsureHelperScripts(t *testing.T) {
	binDir := t.TempDir()
	if err := EnsureHelperScripts(binDir); err != nil {
		t.Fatalf("EnsureHelperScripts: %v", err)
	}
	scriptPath := filepath.Join(binDir, "statusline-helper.sh")
	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("script not written: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Errorf("script is not executable, mode = %o", info.Mode())
	}
	// Idempotent: calling a second time must not error.
	if err := EnsureHelperScripts(binDir); err != nil {
		t.Fatalf("EnsureHelperScripts second call: %v", err)
	}
}
