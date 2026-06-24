package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewWritesToLogFile(t *testing.T) {
	dir := t.TempDir()
	lg, err := New(dir, "info")
	if err != nil {
		t.Fatal(err)
	}
	lg.Info("hello", "k", "v")

	b, err := os.ReadFile(filepath.Join(dir, "logs", "drishti.log"))
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}
	if !strings.Contains(string(b), "hello") {
		t.Errorf("log line not written; got %q", string(b))
	}
}
