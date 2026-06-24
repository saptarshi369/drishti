package security

import (
	"fmt"
	"strings"

	"github.com/saptarshi369/drishti/internal/model"
)

// Evaluate runs every enabled rule over the inputs and returns the findings in
// rule order. It is a pure function — no I/O, no clock — which makes it the
// engine's table-test seam. Secret detection happened upstream in the parsers,
// so only key NAMES reach here; Evaluate never sees a secret value.
func (rs Rules) Evaluate(in model.SecurityInputs) []model.Finding {
	var out []model.Finding
	for _, r := range rs {
		if !r.Enabled {
			continue
		}
		switch r.Type {
		case "require-deny-for-path":
			out = append(out, r.evalRequireDeny(in)...)
		case "forbid-mode":
			out = append(out, r.evalForbidMode(in)...)
		case "broad-allow":
			out = append(out, r.evalBroadAllow(in)...)
		case "untrusted-source":
			out = append(out, r.evalUntrustedSource(in)...)
		case "secret-in-env":
			out = append(out, r.evalSecretInEnv(in)...)
		case "secret-in-settings":
			out = append(out, r.evalSecretInSettings(in)...)
		}
	}
	return out
}

// finding builds a Finding stamped with this rule's presentation fields.
func (r Rule) finding(targetKey, detail, scope string) model.Finding {
	return model.Finding{
		RuleID:      r.ID,
		Severity:    r.Severity,
		Title:       r.Title,
		TargetKey:   targetKey,
		Detail:      detail,
		Remediation: r.Remediation,
		Scope:       scope,
	}
}

// evalRequireDeny emits ONE finding when no deny entry across any scope covers
// any of the rule's path patterns. Matching is a case-insensitive substring
// test (v1 heuristic): a deny like "Read(.env)" covers pattern ".env".
func (r Rule) evalRequireDeny(in model.SecurityInputs) []model.Finding {
	for _, sp := range in.Permissions {
		for _, d := range sp.Deny {
			ld := strings.ToLower(d)
			for _, p := range r.Patterns {
				if strings.Contains(ld, strings.ToLower(p)) {
					return nil // covered somewhere → no finding
				}
			}
		}
	}
	detail := fmt.Sprintf("No permissions.deny rule covers %s.", strings.Join(r.Patterns, " / "))
	return []model.Finding{r.finding("global", detail, "all")}
}

// evalForbidMode emits a finding per scope whose defaultMode is forbidden.
func (r Rule) evalForbidMode(in model.SecurityInputs) []model.Finding {
	var out []model.Finding
	for _, sp := range in.Permissions {
		for _, m := range r.Modes {
			if sp.DefaultMode == m {
				tk := string(sp.Scope) + ":" + sp.RelPath
				detail := fmt.Sprintf("defaultMode is %q in %s.", m, sp.RelPath)
				out = append(out, r.finding(tk, detail, string(sp.Scope)))
			}
		}
	}
	return out
}

// evalBroadAllow emits a finding per allow entry that exactly matches one of the
// rule's too-broad patterns (e.g. "Bash(*)").
func (r Rule) evalBroadAllow(in model.SecurityInputs) []model.Finding {
	// Pre-build a set of forbidden patterns for O(1) lookup per allow entry.
	broad := map[string]bool{}
	for _, p := range r.Patterns {
		broad[p] = true
	}
	var out []model.Finding
	for _, sp := range in.Permissions {
		for _, a := range sp.Allow {
			if broad[a] {
				tk := string(sp.Scope) + ":" + sp.RelPath + ":" + a
				detail := fmt.Sprintf("Allow rule %q in %s is overly broad.", a, sp.RelPath)
				out = append(out, r.finding(tk, detail, string(sp.Scope)))
			}
		}
	}
	return out
}

// evalUntrustedSource emits a finding per plugin whose marketplace is set and is
// not in the allowed list. A plugin with no marketplace is skipped (unknowable).
func (r Rule) evalUntrustedSource(in model.SecurityInputs) []model.Finding {
	// Pre-build a set of allowed marketplace names for O(1) lookup.
	allowed := map[string]bool{}
	for _, a := range r.Allowed {
		allowed[a] = true
	}
	var out []model.Finding
	for _, pl := range in.Plugins {
		if pl.Marketplace == "" || allowed[pl.Marketplace] {
			continue
		}
		tk := "plugin:" + pl.Name
		detail := fmt.Sprintf("Plugin %q comes from untrusted source %q.", pl.Name, pl.Marketplace)
		out = append(out, r.finding(tk, detail, string(pl.Scope)))
	}
	return out
}

// evalSecretInEnv emits a finding per MCP env key already flagged as secret-like.
func (r Rule) evalSecretInEnv(in model.SecurityInputs) []model.Finding {
	var out []model.Finding
	for _, e := range in.MCPEnv {
		for _, k := range e.SecretKeys {
			tk := "mcp:" + e.Server + ":" + k
			detail := fmt.Sprintf("MCP server %q env key %q looks like a secret.", e.Server, k)
			out = append(out, r.finding(tk, detail, string(e.Scope)))
		}
	}
	return out
}

// evalSecretInSettings emits a finding per settings key already flagged as
// secret-like.
func (r Rule) evalSecretInSettings(in model.SecurityInputs) []model.Finding {
	var out []model.Finding
	for _, sp := range in.Permissions {
		for _, k := range sp.SecretSettingKeys {
			tk := string(sp.Scope) + ":" + sp.RelPath + ":" + k
			detail := fmt.Sprintf("Settings key %q in %s looks like a plaintext secret.", k, sp.RelPath)
			out = append(out, r.finding(tk, detail, string(sp.Scope)))
		}
	}
	return out
}
