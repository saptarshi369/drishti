package ingest

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestPollLoopPicksUpNewData(t *testing.T) {
	_, r, dir := setup(t)
	p := filepath.Join(dir, "t.jsonl")
	// User lines must carry string content so the parser counts them as prompts
	// (tool_result-only lines are not prompts; see parse.go isPrompt).
	os.WriteFile(p, []byte(`{"type":"user","timestamp":"2026-06-21T07:00:00Z","sessionId":"s1","message":{"content":"hi"}}`+"\n"), 0o644)

	var inserted int64
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.pollLoop(ctx, 200*time.Millisecond, func(n int) { atomic.AddInt64(&inserted, int64(n)) })

	time.Sleep(300 * time.Millisecond) // first tick ingests the initial line
	f, _ := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString(`{"type":"user","timestamp":"2026-06-21T07:01:00Z","sessionId":"s1","message":{"content":"hi2"}}` + "\n")
	f.Close()
	time.Sleep(400 * time.Millisecond) // next tick ingests the appended line

	if atomic.LoadInt64(&inserted) < 2 {
		t.Errorf("inserted = %d, want >= 2", atomic.LoadInt64(&inserted))
	}
}

// TestWatchPicksUpNestedWrites reproduces the "live activity stops tracking" bug.
// fsnotify watches only the top-level root (e.g. ~/.claude/projects) and is NOT
// recursive, so a transcript written inside a per-project SUBDIRECTORY never fires
// a watcher event. Only a periodic safety-poll inside Watch can catch it. The
// subdir is created BEFORE Watch starts, so the later nested file write produces no
// event for the watched root — meaning a watcher with no poll ingests nothing.
func TestWatchPicksUpNestedWrites(t *testing.T) {
	_, r, dir := setup(t)
	sub := filepath.Join(dir, "proj-x")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	var inserted int64
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Watch(ctx, func(n int) { atomic.AddInt64(&inserted, int64(n)) })

	time.Sleep(150 * time.Millisecond) // let Watch register its watcher
	// Nested transcript write — no fsnotify event reaches the top-level root.
	line := `{"type":"user","timestamp":"2026-06-21T07:00:00Z","sessionId":"s1","message":{"content":"hi"}}` + "\n"
	if err := os.WriteFile(filepath.Join(sub, "t.jsonl"), []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) && atomic.LoadInt64(&inserted) < 1 {
		time.Sleep(50 * time.Millisecond)
	}
	if atomic.LoadInt64(&inserted) < 1 {
		t.Errorf("Watch did not ingest a nested transcript write; inserted = %d, want >= 1",
			atomic.LoadInt64(&inserted))
	}
}

func TestWatchReturnsOnContextCancel(t *testing.T) {
	_, r, _ := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { r.Watch(ctx, nil); close(done) }()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Watch did not return on context cancel")
	}
}
