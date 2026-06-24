package ingest

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"

	"github.com/saptarshi369/drishti/internal/model"
)

// Action is the reconcile ladder's verdict for one file (spec §8.2).
type Action int

const (
	// Skip means the file is unchanged since we last cleanly read it.
	Skip Action = iota
	// Forward means the file grew with the same identity: read from last_offset.
	Forward
	// Reset means rotation/replacement/truncation: re-read from offset 0.
	Reset
	// Missing means stat failed (file gone): keep history, mark state.
	Missing
)

// headHash returns the sha1 hex of the first n bytes of path. It detects a
// same-name file being replaced even when size happens to match.
func headHash(path string, n int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	buf := make([]byte, n)
	read, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", err
	}
	sum := sha1.Sum(buf[:read])
	return fmt.Sprintf("%x", sum), nil
}

// classify applies the cheapest-first ladder. prev is the ledger row; the other
// args are the file's current signals. A brand-new file (prev.Size==0,
// LastOffset==0) classifies as Forward (read from 0).
func classify(prev model.SourceFile, curSize, curMtime int64, ident, head string) Action {
	// Identity or head-hash mismatch => the bytes before our offset changed.
	if prev.FileID != "" && ident != "" && ident != prev.FileID {
		return Reset
	}
	if prev.HeadHash != "" && head != "" && head != prev.HeadHash {
		return Reset
	}
	if curSize < prev.LastOffset {
		return Reset // truncated/shrank below where we'd read
	}
	if curSize == prev.Size && curMtime == prev.MtimeMs && prev.LastOffset == prev.Size {
		return Skip
	}
	return Forward
}
