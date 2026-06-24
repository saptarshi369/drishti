package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileIdentityStableAcrossCalls(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f")
	os.WriteFile(p, []byte("x"), 0o644)

	a, err := FileIdentity(p)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := FileIdentity(p)
	if a == "" || a != b {
		t.Errorf("identity unstable: %q vs %q", a, b)
	}
}

func TestFileIdentityDistinctFiles(t *testing.T) {
	d := t.TempDir()
	p1 := filepath.Join(d, "a")
	p2 := filepath.Join(d, "b")
	os.WriteFile(p1, []byte("1"), 0o644)
	os.WriteFile(p2, []byte("2"), 0o644)
	i1, _ := FileIdentity(p1)
	i2, _ := FileIdentity(p2)
	if i1 == i2 {
		t.Errorf("distinct files share identity %q", i1)
	}
}
