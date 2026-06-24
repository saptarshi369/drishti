package claude

import "testing"

func TestClassifyBlocked(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{
		{"Permission to use Bash has been denied. IMPORTANT: ...", true},
		{"Operation blocked by hook", true},
		{"Over-implementation violation — no failing test output", true},
		{"Exit code 1\ntotal 136\ndrwxr-xr-x ...", false},
		{"<tool_use_error>File has not been read yet.</tool_use_error>", false},
		{"", false},
	}
	for _, c := range cases {
		if got := classifyBlocked(c.text); got != c.want {
			t.Errorf("classifyBlocked(%q) = %v, want %v", c.text, got, c.want)
		}
	}
}
