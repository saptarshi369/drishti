// Package security holds the configurable rule engine for the Security & Audit
// screen. It loads typed rules from a user-editable TOML file and evaluates
// them against parsed config inputs to produce findings. It performs no I/O
// beyond reading the rules file and never sees or stores secret values.
package security

import (
	"bytes"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/saptarshi369/drishti/internal/model"
)

//go:embed security-rules.toml
var defaultRulesTOML []byte

// Rule is one entry in the rules file. Type selects the built-in matcher; the
// remaining parameter fields belong to that matcher (see engine.go).
// The json tags match the toml tags so the Settings API and the TOML file use
// the same field names, keeping the UI and file format consistent.
type Rule struct {
	ID          string   `toml:"id"          json:"id"`
	Type        string   `toml:"type"        json:"type"`
	Enabled     bool     `toml:"enabled"     json:"enabled"`
	Severity    string   `toml:"severity"    json:"severity"`
	Title       string   `toml:"title"       json:"title"`
	Remediation string   `toml:"remediation" json:"remediation"`
	Patterns    []string `toml:"patterns"    json:"patterns,omitempty"`
	Modes       []string `toml:"modes"       json:"modes,omitempty"`
	Allowed     []string `toml:"allowed"     json:"allowed,omitempty"`
	Keywords    []string `toml:"keywords"    json:"keywords,omitempty"`
	Prefixes    []string `toml:"prefixes"    json:"prefixes,omitempty"`
}

// Rules is a loaded, validated rule set.
type Rules []Rule

// rulesFile is the TOML document shape: a list of [[rule]] tables.
type rulesFile struct {
	Rule []Rule `toml:"rule"`
}

// knownTypes is the fixed set of matcher kinds engine.go implements. A rule of
// any other type is dropped at load time.
var knownTypes = map[string]bool{
	"require-deny-for-path": true,
	"forbid-mode":           true,
	"broad-allow":           true,
	"untrusted-source":      true,
	"secret-in-env":         true,
	"secret-in-settings":    true,
}

// parseRules decodes rules TOML and drops any rule with an unknown type or an
// invalid severity, returning a warning string for each drop. One bad rule must
// never disable the whole scan. A document that will not decode returns an error.
func parseRules(data []byte) (Rules, []string, error) {
	var f rulesFile
	// toml.Decode mirrors the project's existing usage in internal/config.
	if _, err := toml.Decode(string(data), &f); err != nil {
		return nil, nil, err
	}
	var out Rules
	var warns []string
	for _, r := range f.Rule {
		if !knownTypes[r.Type] {
			warns = append(warns, fmt.Sprintf("rule %q: unknown type %q (skipped)", r.ID, r.Type))
			continue
		}
		if !model.ValidSeverity(r.Severity) {
			warns = append(warns, fmt.Sprintf("rule %q: invalid severity %q (skipped)", r.ID, r.Severity))
			continue
		}
		out = append(out, r)
	}
	return out, warns, nil
}

// DefaultRules returns the embedded default rule set. The embedded file is
// validated by a unit test, so decoding is not expected to fail; if it somehow
// does, an empty set is returned (degrade, never crash — §14).
func DefaultRules() Rules {
	r, _, err := parseRules(defaultRulesTOML)
	if err != nil {
		return nil
	}
	return r
}

// LoadRulesFromPath reads the user-editable rules file at path. On any problem
// (missing, unreadable, malformed) it logs a warning and falls back to the
// embedded default — the scan must never be disabled by a bad file (§14).
// Per-rule load warnings (unknown type, invalid severity) are only emitted when
// a non-nil logger is supplied; callers that want to surface bad user-edited
// rules to the UI or logs should pass a logger.
func LoadRulesFromPath(path string, lg *slog.Logger) Rules {
	data, err := os.ReadFile(path)
	if err != nil {
		if lg != nil {
			lg.Warn("security rules unreadable; using built-in defaults", "path", path, "err", err)
		}
		return DefaultRules()
	}
	rules, warns, err := parseRules(data)
	if err != nil {
		if lg != nil {
			lg.Warn("security rules malformed; using built-in defaults", "path", path, "err", err)
		}
		return DefaultRules()
	}
	if lg != nil {
		for _, w := range warns {
			lg.Warn("security rules", "note", w)
		}
	}
	return rules
}

// EnsureRulesFile writes the embedded default to path when no file exists there,
// giving the user a documented file to edit. An existing file is left untouched
// so user edits are never clobbered. The parent directory is created
// automatically (mode 0755) so callers need not pre-create it.
func EnsureRulesFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, defaultRulesTOML, 0o644)
}

// WriteRules writes rules to path as TOML (as a [[rule]] array), prefixed with
// a generated header comment that identifies the file and explains that changes
// are picked up automatically within ~10 s by the Drishti daemon scheduler.
// The write is atomic: the encoder fills an in-memory buffer, then the bytes
// are written to a temp file in the same directory and renamed over the target —
// a reader always sees either the old or the new file, never a partial one
// (same pattern as config.Save; see internal/config/config.go).
func WriteRules(path string, rules Rules) error {
	// Header comment: documents the file and mentions the hot-reload cadence so
	// a user who opens the file manually knows it is managed by Drishti Settings.
	const header = "# security-rules.toml — managed by Drishti.\n" +
		"# Edit here or via the Settings UI; changes apply automatically within ~10s.\n\n"

	var buf bytes.Buffer
	buf.WriteString(header)
	// Wrap in rulesFile so the encoder produces the [[rule]] table-array shape
	// that LoadRulesFromPath / parseRules expect. rulesFile is defined in this
	// package so it is directly accessible here.
	if err := toml.NewEncoder(&buf).Encode(rulesFile{Rule: rules}); err != nil {
		return fmt.Errorf("encode rules: %w", err)
	}
	return atomicWriteRulesFile(path, buf.Bytes())
}

// atomicWriteRulesFile writes data to path via a temp file + os.Rename. Using
// the same directory as path ensures src and dst are on the same filesystem,
// which is required for os.Rename to be atomic (POSIX guarantee).
func atomicWriteRulesFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".write-*.tmp")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()        //nolint:errcheck // already in error path
		os.Remove(tmpName) //nolint:errcheck // best-effort cleanup
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName) //nolint:errcheck // best-effort cleanup
		return fmt.Errorf("close temp: %w", err)
	}
	return os.Rename(tmpName, path)
}

// SecretKeywords unions the key-name keywords declared on the secret-* rules.
// An empty result means "use the SecretMatcher's built-in defaults".
func (rs Rules) SecretKeywords() []string {
	return rs.collectSecretField(func(r Rule) []string { return r.Keywords })
}

// SecretPrefixes unions the value prefixes declared on the secret-* rules.
func (rs Rules) SecretPrefixes() []string {
	return rs.collectSecretField(func(r Rule) []string { return r.Prefixes })
}

// collectSecretField iterates over secret-* rules and unions the field selected
// by pick, deduplicating entries while preserving order.
func (rs Rules) collectSecretField(pick func(Rule) []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, r := range rs {
		if r.Type != "secret-in-env" && r.Type != "secret-in-settings" {
			continue
		}
		for _, v := range pick(r) {
			if !seen[v] {
				seen[v] = true
				out = append(out, v)
			}
		}
	}
	return out
}
