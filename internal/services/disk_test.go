package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiskEstimateSumsSizes(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "drishti.db"), make([]byte, 100), 0o644)
	os.WriteFile(filepath.Join(dir, "drishti.db-wal"), make([]byte, 50), 0o644)
	bdir := filepath.Join(dir, "backups")
	os.Mkdir(bdir, 0o755)
	os.WriteFile(filepath.Join(bdir, "old.gz"), make([]byte, 200), 0o644)

	db, bk := DiskEstimate(dir, bdir)
	if db != 150 {
		t.Errorf("db bytes = %d, want 150", db)
	}
	if bk != 200 {
		t.Errorf("backup bytes = %d, want 200", bk)
	}
}

func TestDiskEstimateMissingPathsAreZero(t *testing.T) {
	db, bk := DiskEstimate("/no/such/dir", "/no/such/backups")
	if db != 0 || bk != 0 {
		t.Errorf("missing paths should be 0, got %d %d", db, bk)
	}
}
