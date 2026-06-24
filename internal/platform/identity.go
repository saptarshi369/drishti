// Package platform isolates OS-specific behaviour behind a small interface.
//
// The thin slice targets darwin and linux. The one thing ingest needs from the
// OS is a stable file identity (device + inode) so it can tell "the same file
// grew" from "a new file took this path" (log rotation/replacement). When the
// OS can't supply it, we return an empty id (not an error) and ingest degrades
// to size+mtime+head_hash comparison.
package platform

import (
	"fmt"
	"os"
	"syscall"
)

// FileIdentity returns a stable "device:inode" string for path, or "" if the
// platform does not expose inode data. A non-nil error means the stat itself
// failed (e.g. the file vanished), which the caller treats as MISSING.
func FileIdentity(path string) (string, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return "", nil // unsupported platform: degrade, don't fail
	}
	return fmt.Sprintf("%d:%d", st.Dev, st.Ino), nil
}
