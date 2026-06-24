// Package ingest turns the agent's raw files into plain database rows.
//
// It reads ONLY data we have not already read, using the ledger in the store,
// so on restart we reconcile cheaply instead of re-parsing everything. It must
// never compute "resolved" or "derived" views — it only produces raw facts.
package ingest

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/saptarshi369/drishti/internal/model"
	claude "github.com/saptarshi369/drishti/internal/sources/claude"
	"github.com/saptarshi369/drishti/internal/store"
)

// headBytes is how many leading bytes feed head_hash (rotation detection).
const headBytes = 4096

// Reconciler walks transcript roots and applies the incremental read ladder.
type Reconciler struct {
	st    *store.Store
	roots []string
	lg    *slog.Logger
	ident func(string) (string, error) // injectable for tests; defaults to platform
}

// New constructs a Reconciler over the given roots. lg may be nil.
func New(st *store.Store, roots []string, lg *slog.Logger) *Reconciler {
	if lg == nil {
		lg = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Reconciler{st: st, roots: roots, lg: lg, ident: fileIdentity}
}

// ledgerByPath loads the existing ledger row for path, or a zero row if new.
func (r *Reconciler) ledgerByPath(path string) (model.SourceFile, error) {
	files, err := r.st.ListSourceFiles()
	if err != nil {
		return model.SourceFile{}, err
	}
	for _, f := range files {
		if f.AbsPath == path {
			return f, nil
		}
	}
	return model.SourceFile{AbsPath: path, AgentCode: "claude", Kind: "transcript"}, nil
}

// ReconcileFile runs the ladder for a single file and returns new events inserted.
func (r *Reconciler) ReconcileFile(path string) (int, error) {
	prev, err := r.ledgerByPath(path)
	if err != nil {
		return 0, err
	}

	fi, statErr := os.Stat(path)
	if statErr != nil {
		// MISSING: keep historical rows, just mark state.
		prev.State = "missing"
		_, _ = r.st.UpsertSourceFile(prev)
		r.lg.Debug("source missing", "path", path)
		return 0, nil
	}

	ident, _ := r.ident(path)
	head, _ := headHash(path, headBytes)
	action := classify(prev, fi.Size(), fi.ModTime().UnixMilli(), ident, head)

	if action == Skip {
		return 0, nil
	}

	startOffset := prev.LastOffset
	if action == Reset {
		startOffset = 0
	}
	return r.readFrom(path, prev, fi, ident, head, startOffset)
}

// readFrom opens path, seeks to startOffset, parses the remainder, and commits
// rows + the advanced offset in one transaction (via store.ApplyIngest).
func (r *Reconciler) readFrom(path string, prev model.SourceFile, fi os.FileInfo, ident, head string, startOffset int64) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Seek(startOffset, io.SeekStart); err != nil {
		return 0, err
	}

	res, err := claude.Parse(f, claude.ParseContext{})
	if err != nil {
		return 0, err
	}

	// Persist the ledger row first so we have its id, carrying the NEW identity.
	prev.FileID, prev.HeadHash = ident, head
	prev.Size, prev.MtimeMs = fi.Size(), fi.ModTime().UnixMilli()
	prev.State = "active"
	sfID, err := r.st.UpsertSourceFile(prev)
	if err != nil {
		return 0, err
	}

	newOffset := startOffset + res.BytesConsumed
	inserted, err := r.st.ApplyIngest(store.IngestBatch{
		SourceFileID: sfID,
		ProjectRoot:  projectKey(path),
		Events:       res.Events,
		Deltas:       res.Deltas,
		NewOffset:    newOffset,
		NewLine:      prev.LastLine + res.LinesConsumed,
		ReadMs:       time.Now().UnixMilli(),
	})
	if err != nil {
		return 0, err
	}
	if res.ErrorCount > 0 {
		r.lg.Debug("skipped malformed lines", "path", path, "count", res.ErrorCount)
	}
	return inserted, nil
}

// projectKey returns the encoded Claude "projects" directory that contains a
// transcript file — i.e. filepath.Base(filepath.Dir(path)). Claude stores each
// project's transcripts under ~/.claude/projects/<encoded-root>/, so this dir
// name is a stable per-project grouping key for usage attribution. We do NOT try
// to decode it back to an absolute path (the encoding is lossy); the display
// label is derived from it later in the services layer.
func projectKey(path string) string {
	return filepath.Base(filepath.Dir(path))
}

// walkJSONL invokes fn for every *.jsonl file under root.
func walkJSONL(root string, fn func(path string)) error {
	return filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(p) != ".jsonl" {
			return nil
		}
		fn(p)
		return nil
	})
}

// ScanAll globs the roots for transcript files and reconciles each (used by the
// startup ladder and the poll fallback).
func (r *Reconciler) ScanAll() error {
	for _, root := range r.roots {
		if err := walkJSONL(root, func(p string) {
			if _, e := r.ReconcileFile(p); e != nil {
				r.lg.Warn("reconcile failed", "path", p, "err", e)
			}
		}); err != nil {
			return err
		}
	}
	return nil
}
