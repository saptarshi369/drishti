package ingest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/store"
)

// setup builds a fresh store + Reconciler over a temp dir root.
func setup(t *testing.T) (*store.Store, *Reconciler, string) {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "drishti.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	r := New(st, []string{dir}, nil)
	return st, r, dir
}

func TestClassifyDecisions(t *testing.T) {
	prev := model.SourceFile{Size: 100, MtimeMs: 5, FileID: "1:2", HeadHash: "h", LastOffset: 100}
	cases := []struct {
		name        string
		curSize     int64
		curMtime    int64
		ident, head string
		want        Action
	}{
		{"unchanged", 100, 5, "1:2", "h", Skip},
		{"grew", 180, 6, "1:2", "h", Forward},
		{"shrank", 40, 7, "1:2", "h", Reset},
		{"identity-changed", 180, 6, "9:9", "h", Reset},
		{"headhash-changed", 180, 6, "1:2", "DIFF", Reset},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := classify(prev, c.curSize, c.curMtime, c.ident, c.head)
			if got != c.want {
				t.Errorf("classify(%s) = %v, want %v", c.name, got, c.want)
			}
		})
	}
}

func TestReconcileForwardThenUnchanged(t *testing.T) {
	_, r, dir := setup(t)
	p := filepath.Join(dir, "t.jsonl")
	// User lines must carry string content so the parser counts them as prompts
	// (tool_result-only lines are not prompts; see parse.go isPrompt).
	os.WriteFile(p, []byte(`{"type":"user","timestamp":"2026-06-21T07:00:00Z","sessionId":"s1","message":{"content":"hi"}}`+"\n"), 0o644)

	n, err := r.ReconcileFile(p)
	if err != nil || n != 1 {
		t.Fatalf("first reconcile inserted=%d err=%v, want 1", n, err)
	}
	n2, _ := r.ReconcileFile(p)
	if n2 != 0 {
		t.Errorf("unchanged reconcile inserted=%d, want 0", n2)
	}
	f, _ := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString(`{"type":"user","timestamp":"2026-06-21T07:01:00Z","sessionId":"s1","message":{"content":"hi2"}}` + "\n")
	f.Close()
	n3, _ := r.ReconcileFile(p)
	if n3 != 1 {
		t.Errorf("forward reconcile inserted=%d, want 1", n3)
	}
}

func TestReconcileRotationResets(t *testing.T) {
	_, r, dir := setup(t)
	p := filepath.Join(dir, "t.jsonl")
	// User lines must carry string content so the parser counts them as prompts
	// (tool_result-only lines are not prompts; see parse.go isPrompt).
	os.WriteFile(p, []byte(`{"type":"user","timestamp":"2026-06-21T07:00:00Z","sessionId":"s1","message":{"content":"hi"}}`+"\n"), 0o644)
	r.ReconcileFile(p)
	// Replace with a smaller, different file (rotation): shrank → Reset → re-read.
	os.WriteFile(p, []byte(`{"type":"user","timestamp":"2026-06-21T08:00:00Z","sessionId":"s2","message":{"content":"hey"}}`+"\n"), 0o644)
	n, err := r.ReconcileFile(p)
	if err != nil || n != 1 {
		t.Errorf("reset reconcile inserted=%d err=%v, want 1", n, err)
	}
}

func TestReconcileMissingMarksState(t *testing.T) {
	st, r, dir := setup(t)
	p := filepath.Join(dir, "t.jsonl")
	os.WriteFile(p, []byte(`{"type":"user","timestamp":"2026-06-21T07:00:00Z","sessionId":"s1","message":{"content":"hi"}}`+"\n"), 0o644)
	r.ReconcileFile(p)
	os.Remove(p)
	if _, err := r.ReconcileFile(p); err != nil {
		t.Errorf("missing file must not error, got %v", err)
	}
	files, _ := st.ListSourceFiles()
	if files[0].State != "missing" {
		t.Errorf("state = %q, want missing", files[0].State)
	}
}

// TestProjectKey extracts the encoded Claude projects directory from a transcript
// path. It is the per-project grouping key for usage_rollup attribution.
func TestProjectKey(t *testing.T) {
	cases := []struct{ path, want string }{
		{"/home/me/.claude/projects/-Users-me-dev-myapp/abc.jsonl", "-Users-me-dev-myapp"},
		{"/x/-Users-me-payments-svc/s.jsonl", "-Users-me-payments-svc"},
		{"plain.jsonl", "."}, // filepath.Dir("plain.jsonl") == ".", Base(".") == "."
	}
	for _, c := range cases {
		if got := projectKey(c.path); got != c.want {
			t.Errorf("projectKey(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestReconcileCrashReplayIdempotent(t *testing.T) {
	st, r, dir := setup(t)
	p := filepath.Join(dir, "t.jsonl")
	os.WriteFile(p, []byte(`{"type":"user","timestamp":"2026-06-21T07:00:00Z","sessionId":"s1","message":{"content":"hi"}}`+"\n"), 0o644)
	r.ReconcileFile(p)
	// Simulate a crash before the offset advanced: rewind the ledger offset to 0,
	// then reconcile again. dedupe_key makes re-presented lines a no-op.
	st.DB().Exec("UPDATE source_files SET last_offset=0, last_line=0")
	n, _ := r.ReconcileFile(p)
	if n != 0 {
		t.Errorf("replay inserted=%d, want 0 (idempotent)", n)
	}
}
