package services

// disk.go — Module 7 retention disk estimate.
// DiskEstimate provides a best-effort gauge of on-disk footprint (§14: no panics, no errors).

import (
	"os"
	"path/filepath"
	"strings"
)

// DiskEstimate returns the approximate on-disk footprint for the Retention
// panel. It sums the sizes of SQLite database files (any file whose name
// contains ".db", e.g. drishti.db, drishti.db-wal, drishti.db-shm) found
// directly in dataDir, and recursively sums all files under backupDir.
//
// Missing or unreadable paths contribute 0 — this function is a best-effort
// gauge and never returns an error (Drishti §14 failsafe: degrade, don't die).
func DiskEstimate(dataDir, backupDir string) (dbBytes, backupBytes int64) {
	// Read the top-level entries in dataDir. If the directory does not exist
	// or is unreadable, ReadDir returns an error and we silently contribute 0.
	if entries, err := os.ReadDir(dataDir); err == nil {
		for _, e := range entries {
			// Skip sub-directories; we only want the flat DB files here.
			if e.IsDir() {
				continue
			}
			// Match *.db / *.db-wal / *.db-shm (and any other .db* suffixes)
			// by checking whether the filename contains the substring ".db".
			if strings.Contains(e.Name(), ".db") {
				// e.Info() re-stats the entry; treat any error as 0.
				if info, err := e.Info(); err == nil {
					dbBytes += info.Size()
				}
			}
		}
	}

	// Walk the entire backup directory tree. WalkDir visits every file and
	// sub-directory; we skip directories (IsDir) and silently skip errors.
	//nolint:errcheck // best-effort: walk errors are deliberately ignored (§14)
	filepath.WalkDir(backupDir, func(_ string, d os.DirEntry, err error) error {
		// err is set when the entry itself cannot be stat'd; skip it.
		if err != nil || d.IsDir() {
			return nil
		}
		if info, err := d.Info(); err == nil {
			backupBytes += info.Size()
		}
		return nil
	})

	return dbBytes, backupBytes
}
