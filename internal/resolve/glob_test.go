package resolve

import "testing"

func TestGlobMatch(t *testing.T) {
	cases := []struct {
		pattern, name string
		want          bool
	}{
		{"**/CLAUDE.md", "/a/b/CLAUDE.md", true},
		{"**/CLAUDE.md", "/CLAUDE.md", true},
		{"/home/**", "/home/x/y", true},
		{"*.md", "a.md", true},
		{"*.md", "a/b.md", false}, // * does not cross a separator
		{"a?c", "abc", true},
		{"a?c", "a/c", false},
		{"**/other-team/CLAUDE.md", "/repo/other-team/CLAUDE.md", true},
		{"**/other-team/CLAUDE.md", "/repo/my-team/CLAUDE.md", false},
		{"exact", "exact", true},
		{"exact", "other", false},
	}
	for _, c := range cases {
		if got := globMatch(c.pattern, c.name); got != c.want {
			t.Errorf("globMatch(%q, %q) = %v, want %v", c.pattern, c.name, got, c.want)
		}
	}
}
