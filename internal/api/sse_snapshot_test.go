package api

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/store"
)

func TestStreamSendsSnapshotOnConnect(t *testing.T) {
	st, _ := store.Open(filepath.Join(t.TempDir(), "drishti.db"))
	defer st.Close()
	srv := NewServer("test", st)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/api/stream", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// The handler must push an initial snapshot frame without any broadcast.
	line, err := bufio.NewReader(resp.Body).ReadString('\n')
	if err != nil || !strings.HasPrefix(line, "data: ") {
		t.Fatalf("expected an immediate data frame, got %q (err %v)", line, err)
	}
}

func TestBroadcastSnapshotSendsCounters(t *testing.T) {
	st, _ := store.Open(filepath.Join(t.TempDir(), "drishti.db"))
	defer st.Close()
	srv := NewServer("test", st)
	ch, cancel := srv.Hub().Subscribe()
	defer cancel()

	srv.BroadcastSnapshot()

	select {
	case m := <-ch:
		if m.Type != "counters" {
			t.Errorf("first broadcast type = %q, want counters", m.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("BroadcastSnapshot sent nothing")
	}
}

func TestCountersFrameCarriesM8Fields(t *testing.T) {
	st, _ := store.Open(filepath.Join(t.TempDir(), "drishti.db"))
	defer st.Close()
	srv := NewServer("test", st)
	ch, cancel := srv.Hub().Subscribe()
	defer cancel()

	srv.BroadcastSnapshot()

	select {
	case m := <-ch:
		if m.Type != "counters" {
			t.Fatalf("first frame = %q, want counters", m.Type)
		}
		b, err := json.Marshal(m.Payload)
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{`"active_components"`, `"health"`, `"alerts"`, `"context_tax"`} {
			if !strings.Contains(string(b), key) {
				t.Errorf("counters payload missing %s; got %s", key, b)
			}
		}
	case <-time.After(time.Second):
		t.Fatal("BroadcastSnapshot sent nothing")
	}
}

// TestSnapshotMessagesIncludesActivity verifies that snapshotMessages returns
// an "activity" message in addition to the existing "counters" and "status"
// messages. It seeds the store with one event so ActivitySnapshot has data to
// work with, then iterates the returned slice looking for Type == "activity".
func TestSnapshotMessagesIncludesActivity(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/t.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })

	_, err = st.ApplyIngest(store.IngestBatch{Events: []model.Event{
		{
			AgentCode:  "claude",
			TypeCode:   "skill",
			SourceCode: "transcript",
			SessionID:  "s1",
			TsMs:       1000,
			SkillName:  "deploy",
			DedupeKey:  "k1",
		},
	}})
	if err != nil {
		t.Fatal(err)
	}

	srv := NewServer("test", st)
	var hasActivity bool
	for _, m := range srv.snapshotMessages() {
		if m.Type == "activity" {
			hasActivity = true
		}
	}
	if !hasActivity {
		t.Fatal("snapshotMessages missing activity message")
	}
}
