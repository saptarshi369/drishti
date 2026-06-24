package resolve

import "strings"

// globMatch reports whether name matches a glob pattern over the WHOLE string.
// It supports three wildcards, matching Claude Code's claudeMdExcludes globbing:
//   - "?"  matches exactly one character that is not a path separator ('/').
//   - "*"  matches zero or more characters, but never crosses a '/'.
//   - "**" matches zero or more characters INCLUDING '/'.
//
// The implementation is a small recursive backtracker. Recursion is bounded by
// the pattern + name length; paths are short, so the worst-case cost is fine and
// we avoid pulling in a third-party doublestar dependency (tiny-deps rule).
func globMatch(pattern, name string) bool {
	switch {
	case pattern == "":
		// An empty pattern only matches an empty remaining name.
		return name == ""
	case strings.HasPrefix(pattern, "**"):
		// "**" can consume any suffix of name (including '/'), so try matching the
		// rest of the pattern at every position and succeed if any works.
		rest := pattern[2:]
		for i := 0; i <= len(name); i++ {
			if globMatch(rest, name[i:]) {
				return true
			}
		}
		return false
	case pattern[0] == '*':
		// "*" behaves like "**" but stops at the first '/': it cannot span dirs.
		rest := pattern[1:]
		for i := 0; i <= len(name); i++ {
			if globMatch(rest, name[i:]) {
				return true
			}
			if i < len(name) && name[i] == '/' {
				break
			}
		}
		return false
	case pattern[0] == '?':
		// "?" consumes exactly one non-separator character.
		if name == "" || name[0] == '/' {
			return false
		}
		return globMatch(pattern[1:], name[1:])
	default:
		// Literal character: must match the next byte exactly.
		if name == "" || name[0] != pattern[0] {
			return false
		}
		return globMatch(pattern[1:], name[1:])
	}
}
