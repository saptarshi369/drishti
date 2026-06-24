package claude

import "strings"

// DefaultSecretKeywords are env/setting KEY-NAME fragments that imply the value
// is a credential. Matching is case-insensitive substring.
var DefaultSecretKeywords = []string{
	"token", "secret", "password", "passwd", "apikey", "api_key",
	"auth", "credential", "private_key",
}

// DefaultSecretPrefixes are VALUE prefixes that identify well-known key formats.
var DefaultSecretPrefixes = []string{
	"sk-", "ghp_", "gho_", "github_pat_", "AKIA", "xoxb-", "xoxp-", "AIza",
}

// SecretMatcher decides whether a (key, value) pair looks like a secret. The
// lists come from the security rules file; empty lists fall back to the
// package defaults. The value is only ever inspected here — callers keep the
// key name and a boolean, never the value (privacy default D8).
type SecretMatcher struct {
	Keywords []string
	Prefixes []string
}

// Match reports whether key/value looks like a secret. Two arms:
//   - value-prefix: a well-known format (sk-, ghp_, …) is always a secret.
//   - key-keyword: a secret-ish KEY name AND an opaque-looking value. Requiring
//     opacity stops command paths (e.g. the apiKeyHelper script) and prose from
//     tripping the keyword arm.
func (m SecretMatcher) Match(key, value string) bool {
	pfx := m.Prefixes
	if len(pfx) == 0 {
		pfx = DefaultSecretPrefixes
	}
	for _, p := range pfx {
		if strings.HasPrefix(value, p) {
			return true
		}
	}
	if looksOpaque(value) {
		kws := m.Keywords
		if len(kws) == 0 {
			kws = DefaultSecretKeywords
		}
		lk := strings.ToLower(key)
		for _, k := range kws {
			if strings.Contains(lk, strings.ToLower(k)) {
				return true
			}
		}
	}
	return false
}

// looksOpaque reports whether v resembles a raw credential: reasonably long and
// free of whitespace and path separators (so command paths and prose don't trip
// the keyword arm of Match).
func looksOpaque(v string) bool {
	if len(v) < 16 {
		return false
	}
	return !strings.ContainsAny(v, " /\\\t\n")
}
