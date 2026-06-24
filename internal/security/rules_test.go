package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultRules_LoadsSevenValidRules(t *testing.T) {
	r := DefaultRules()
	if len(r) != 7 {
		t.Fatalf("DefaultRules() len = %d, want 7", len(r))
	}
}

func TestParseRules_SkipsUnknownTypeAndSeverity(t *testing.T) {
	doc := `
[[rule]]
id = "ok"
type = "forbid-mode"
enabled = true
severity = "high"
modes = ["bypassPermissions"]

[[rule]]
id = "bad-type"
type = "frobnicate"
enabled = true
severity = "high"

[[rule]]
id = "bad-sev"
type = "forbid-mode"
enabled = true
severity = "spicy"
`
	rules, warns, err := parseRules([]byte(doc))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 1 || rules[0].ID != "ok" {
		t.Fatalf("got %d rules %+v, want only the valid one", len(rules), rules)
	}
	if len(warns) != 2 {
		t.Fatalf("warns = %d, want 2", len(warns))
	}
}

func TestParseRules_MalformedReturnsError(t *testing.T) {
	if _, _, err := parseRules([]byte("this = = not toml")); err == nil {
		t.Fatal("expected error for malformed TOML")
	}
}

func TestEnsureRulesFile_WritesWhenAbsentKeepsWhenPresent(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "security-rules.toml")
	if err := EnsureRulesFile(p); err != nil {
		t.Fatalf("EnsureRulesFile: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("file not written: %v", err)
	}
	// Overwrite with a marker, then ensure again: must NOT clobber user edits.
	if err := os.WriteFile(p, []byte("# edited"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureRulesFile(p); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if string(b) != "# edited" {
		t.Fatal("EnsureRulesFile overwrote an existing file")
	}
}

func TestLoadRulesFromPath_MissingFallsBackToDefault(t *testing.T) {
	r := LoadRulesFromPath(filepath.Join(t.TempDir(), "nope.toml"), nil)
	if len(r) != 7 {
		t.Fatalf("fallback len = %d, want 7", len(r))
	}
}

func TestEnsureRulesFile_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "nested", "deeper", "security-rules.toml")
	if err := EnsureRulesFile(p); err != nil {
		t.Fatalf("EnsureRulesFile: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("file not written under created dirs: %v", err)
	}
}

func TestSecretKeywordsAndPrefixes_UnionDeduplicates(t *testing.T) {
	r := DefaultRules()
	kw := r.SecretKeywords()
	if len(kw) == 0 {
		t.Fatal("SecretKeywords() returned empty")
	}
	pf := r.SecretPrefixes()
	if len(pf) == 0 {
		t.Fatal("SecretPrefixes() returned empty")
	}
	// Both secret-in-env and secret-in-settings share "token" — must appear once.
	count := 0
	for _, k := range kw {
		if k == "token" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("'token' appears %d times in SecretKeywords(), want 1 (dedup)", count)
	}
}
