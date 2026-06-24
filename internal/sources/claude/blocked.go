package claude

import "strings"

// blockedPatterns are case-insensitive substrings that mark a tool_result error
// as a DENIAL/BLOCK (permission refusal or a hook veto) rather than a plain tool
// failure. Derived from real Claude Code transcripts; extend as new block shapes
// are observed. Kept deliberately small and specific to avoid false positives.
var blockedPatterns = []string{
	"permission to use",
	"permission denied",
	"blocked by",
	"has been denied",
	"violation —", // TDD-Guard hook vetoes, e.g. "Over-implementation violation —"
	"operation not permitted by",
}

// classifyBlocked reports whether an is_error tool_result's text is a block/denial
// (vs a generic tool error). Pure and table-tested; reads text transiently — the
// caller stores NONE of it (privacy D8).
func classifyBlocked(text string) bool {
	low := strings.ToLower(text)
	for _, p := range blockedPatterns {
		if strings.Contains(low, p) {
			return true
		}
	}
	return false
}
