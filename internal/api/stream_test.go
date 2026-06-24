package api

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStreamRelaysBroadcast(t *testing.T) {
	srv := NewServer("test", nil)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start broadcasting BEFORE Do(): an SSE client's Do() returns only once the
	// server writes the first frame, so the broadcaster must already be running.
	// The delay gives the handler time to subscribe (Broadcast is non-blocking).
	go func() {
		time.Sleep(100 * time.Millisecond)
		srv.Hub().Broadcast(Message{Type: "counters", TS: 1, Payload: map[string]int{"n": 1}})
	}()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/api/stream", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content-type = %q, want text/event-stream", ct)
	}

	line, err := bufio.NewReader(resp.Body).ReadString('\n')
	if err != nil || !strings.HasPrefix(line, "data: ") {
		t.Fatalf("expected a data frame, got %q (err %v)", line, err)
	}
}
