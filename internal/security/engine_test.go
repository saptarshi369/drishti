package security

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

// rule builds a test Rule with sensible defaults and an optional fields func to
// set type-specific parameters without repetition in each test.
func rule(id, typ string, fields func(*Rule)) Rule {
	r := Rule{ID: id, Type: typ, Enabled: true, Severity: "high", Title: id, Remediation: "fix it"}
	if fields != nil {
		fields(&r)
	}
	return r
}

func TestEvaluate_RequireDeny_MissingEmitsOne(t *testing.T) {
	rules := Rules{rule("missing-env-deny", "require-deny-for-path", func(r *Rule) { r.Patterns = []string{".env"} })}
	in := model.SecurityInputs{Permissions: []model.ScopePermissions{{Scope: "user", RelPath: "settings.json", Deny: []string{"Read(secrets.txt)"}}}}
	got := rules.Evaluate(in)
	if len(got) != 1 || got[0].RuleID != "missing-env-deny" || got[0].TargetKey != "global" {
		t.Fatalf("got %+v, want one global missing-env-deny finding", got)
	}
}

func TestEvaluate_RequireDeny_CoveredEmitsNone(t *testing.T) {
	rules := Rules{rule("missing-env-deny", "require-deny-for-path", func(r *Rule) { r.Patterns = []string{".env"} })}
	in := model.SecurityInputs{Permissions: []model.ScopePermissions{{Deny: []string{"Read(.env)"}}}}
	if got := rules.Evaluate(in); len(got) != 0 {
		t.Fatalf("got %+v, want none (deny present)", got)
	}
}

func TestEvaluate_ForbidMode(t *testing.T) {
	rules := Rules{rule("bypass", "forbid-mode", func(r *Rule) { r.Modes = []string{"bypassPermissions"} })}
	in := model.SecurityInputs{Permissions: []model.ScopePermissions{{Scope: "project", RelPath: "settings.json", DefaultMode: "bypassPermissions"}}}
	got := rules.Evaluate(in)
	if len(got) != 1 || got[0].TargetKey != "project:settings.json" {
		t.Fatalf("got %+v, want one forbid-mode finding", got)
	}
}

func TestEvaluate_BroadAllow(t *testing.T) {
	rules := Rules{rule("broad", "broad-allow", func(r *Rule) { r.Patterns = []string{"Bash(*)"} })}
	in := model.SecurityInputs{Permissions: []model.ScopePermissions{{Scope: "user", RelPath: "settings.json", Allow: []string{"Bash(*)", "Read(foo.txt)"}}}}
	got := rules.Evaluate(in)
	// The broad-allow target_key is scope+":"+relPath+":"+allow — verify count and exact key.
	if len(got) != 1 || got[0].TargetKey != "user:settings.json:Bash(*)" {
		t.Fatalf("got %+v, want one broad-allow finding with target_key %q", got, "user:settings.json:Bash(*)")
	}
}

func TestEvaluate_UntrustedSource(t *testing.T) {
	rules := Rules{rule("untrusted", "untrusted-source", func(r *Rule) { r.Allowed = []string{"anthropics"} })}
	in := model.SecurityInputs{Plugins: []model.PluginSource{
		{Name: "good", Marketplace: "anthropics"},
		{Name: "risky", Marketplace: "rando"},
	}}
	got := rules.Evaluate(in)
	if len(got) != 1 || got[0].TargetKey != "plugin:risky" {
		t.Fatalf("got %+v, want one untrusted finding for risky", got)
	}
}

func TestEvaluate_SecretInEnv(t *testing.T) {
	rules := Rules{rule("mcp-secret", "secret-in-env", nil)}
	in := model.SecurityInputs{MCPEnv: []model.MCPEnvShape{{Server: "db", Scope: "project", SecretKeys: []string{"API_KEY"}}}}
	got := rules.Evaluate(in)
	if len(got) != 1 || got[0].TargetKey != "mcp:db:API_KEY" {
		t.Fatalf("got %+v, want one secret-in-env finding", got)
	}
}

func TestEvaluate_SecretInSettings(t *testing.T) {
	rules := Rules{rule("settings-secret", "secret-in-settings", nil)}
	in := model.SecurityInputs{Permissions: []model.ScopePermissions{{Scope: "user", RelPath: "settings.json", SecretSettingKeys: []string{"anthropicApiKey"}}}}
	got := rules.Evaluate(in)
	// The secret-in-settings target_key is scope+":"+relPath+":"+key.
	if len(got) != 1 || got[0].TargetKey != "user:settings.json:anthropicApiKey" {
		t.Fatalf("got %+v, want one secret-in-settings finding with target_key %q", got, "user:settings.json:anthropicApiKey")
	}
}

func TestEvaluate_DisabledRuleSkipped(t *testing.T) {
	r := rule("bypass", "forbid-mode", func(r *Rule) { r.Modes = []string{"bypassPermissions"} })
	r.Enabled = false
	in := model.SecurityInputs{Permissions: []model.ScopePermissions{{DefaultMode: "bypassPermissions"}}}
	if got := (Rules{r}).Evaluate(in); len(got) != 0 {
		t.Fatalf("disabled rule produced findings: %+v", got)
	}
}
