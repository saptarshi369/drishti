package claude

import (
	"errors"
	"testing"
)

type fmProbe struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tools       []string `yaml:"allowed-tools"`
}

func TestParseFrontmatter(t *testing.T) {
	body := []byte("---\nname: commit\ndescription: Stage and write a commit\nallowed-tools: [Bash, Read]\n---\n# Body text\n")
	var p fmProbe
	if err := parseFrontmatter(body, &p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "commit" {
		t.Errorf("name = %q, want commit", p.Name)
	}
	if p.Description != "Stage and write a commit" {
		t.Errorf("description = %q", p.Description)
	}
	if len(p.Tools) != 2 || p.Tools[0] != "Bash" {
		t.Errorf("tools = %v, want [Bash Read]", p.Tools)
	}
}

func TestParseFrontmatter_NoBlock(t *testing.T) {
	var p fmProbe
	err := parseFrontmatter([]byte("# just markdown\n"), &p)
	if err == nil {
		t.Fatal("want error when no frontmatter block present")
	}
	if !errors.Is(err, errNoFrontmatter) {
		t.Fatalf("want errNoFrontmatter, got %v", err)
	}
}

func TestParseFrontmatter_CRLF(t *testing.T) {
	body := []byte("---\r\nname: commit\r\ndescription: Stage and write a commit\r\nallowed-tools: [Bash, Read]\r\n---\r\n# Body\r\n")
	var p fmProbe
	if err := parseFrontmatter(body, &p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "commit" {
		t.Errorf("name = %q, want commit", p.Name)
	}
	if p.Description != "Stage and write a commit" {
		t.Errorf("description = %q", p.Description)
	}
	if len(p.Tools) != 2 || p.Tools[0] != "Bash" {
		t.Errorf("tools = %v, want [Bash Read]", p.Tools)
	}
}

func TestParseFrontmatter_TightFenceCheck(t *testing.T) {
	// ---yaml should not be a valid opening; only ---\n is valid
	body := []byte("---yaml\nname: test\n---\n")
	var p fmProbe
	err := parseFrontmatter(body, &p)
	if !errors.Is(err, errNoFrontmatter) {
		t.Fatalf("want errNoFrontmatter for ---yaml, got %v", err)
	}
}

func TestParseFrontmatter_BOM(t *testing.T) {
	// Prepend UTF-8 BOM to a valid frontmatter block; should parse correctly.
	body := []byte("\xef\xbb\xbfname: commit\ndescription: Stage and write a commit\nallowed-tools: [Bash, Read]\n---\n# Body text\n")
	body = append([]byte("---\n"), body...)
	var p fmProbe
	if err := parseFrontmatter(body, &p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "commit" {
		t.Errorf("name = %q, want commit", p.Name)
	}
	if p.Description != "Stage and write a commit" {
		t.Errorf("description = %q", p.Description)
	}
	if len(p.Tools) != 2 || p.Tools[0] != "Bash" {
		t.Errorf("tools = %v, want [Bash Read]", p.Tools)
	}
}
