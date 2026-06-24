package app

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSmokeBootServeShutdown(t *testing.T) {
	// Fixture HOME with one transcript so the scan has something to ingest.
	home := t.TempDir()
	projDir := filepath.Join(home, ".claude", "projects", "proj")
	os.MkdirAll(projDir, 0o755)
	os.WriteFile(filepath.Join(projDir, "t.jsonl"),
		[]byte(`{"type":"user","timestamp":"2026-06-21T07:00:00Z","sessionId":"s1","isMeta":false}`+"\n"), 0o644)
	t.Setenv("HOME", home)
	t.Setenv("DRISHTI_DATA_DIR", filepath.Join(home, ".drishti"))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Run(ctx, "smoke") }()

	base := "http://127.0.0.1:7777"
	var up bool
	for i := 0; i < 30; i++ {
		if resp, err := http.Get(base + "/api/health"); err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			up = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !up {
		cancel()
		t.Fatal("server never became healthy")
	}

	resp, err := http.Get(base + "/api/overview")
	if err != nil || resp.StatusCode != 200 {
		cancel()
		t.Fatalf("overview failed: %v status=%v", err, resp)
	}
	resp.Body.Close()

	cancel() // trigger graceful shutdown
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error on shutdown: %v", err)
		}
	case <-time.After(12 * time.Second):
		t.Fatal("daemon did not shut down within deadline")
	}
}
