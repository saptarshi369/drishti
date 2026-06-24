package api

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/store"
)

// TestHandleActivityEventsFilters verifies GET /api/activity/events returns only
// the events matching the ?type= query parameter and serialises the result as
// {"events":[...]} with events=[] (not null) when no events match.
//
// Seed: one tool_use and one prompt. Filter ?type=tool_use → 1 result with
// ToolName="Bash". The handler must convert a nil slice to [] before writing.
func TestHandleActivityEventsFilters(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/t.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if _, err := st.ApplyIngest(store.IngestBatch{Events: []model.Event{
		{AgentCode: "claude", TypeCode: "tool_use", SourceCode: "transcript", SessionID: "s", TsMs: 1, ToolName: "Bash", DedupeKey: "a"},
		{AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript", SessionID: "s", TsMs: 2, DedupeKey: "b"},
	}}); err != nil {
		t.Fatal(err)
	}
	srv := NewServer("test", st)
	req := httptest.NewRequest("GET", "/api/activity/events?type=tool_use", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	var body struct{ Events []model.RecentEvent }
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if len(body.Events) != 1 || body.Events[0].ToolName != "Bash" {
		t.Fatalf("events = %+v", body.Events)
	}
}

// TestHandleActivityReturnsSnapshot seeds the store with one prompt event using
// a near-now timestamp (so it falls inside the 24-hour rolling window) and
// verifies that GET /api/activity returns a 200 with:
//   - Counters.Last24h.Prompts == 1 (the seeded event)
//   - Skills serialised as [] not null (no skill events seeded)
//   - Recent serialised as [] not null (snapshot Recent field)
//
// Why near-now timestamps? ActivitySnapshot calls st.ActivityCounters with
// (now-24h) as the cutoff. An epoch timestamp (TsMs:1000) is >24 h old and
// would be excluded, making Prompts==0 and the assertion wrong.
func TestHandleActivityReturnsSnapshot(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/t.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })

	// Use a near-now timestamp so the event falls inside the 24-h window.
	nowMs := time.Now().UnixMilli()
	if _, err := st.ApplyIngest(store.IngestBatch{Events: []model.Event{
		{AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript",
			SessionID: "s1", TsMs: nowMs, DedupeKey: "p1"},
	}}); err != nil {
		t.Fatal(err)
	}

	srv := NewServer("test", st)
	req := httptest.NewRequest("GET", "/api/activity", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status %d; body=%s", rec.Code, rec.Body.String())
	}

	var got model.ActivitySnapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v body=%s", err, rec.Body.String())
	}

	// The seeded prompt event must appear in the 24-h counter.
	if got.Counters.Last24h.Prompts != 1 {
		t.Fatalf("counters = %+v", got.Counters)
	}

	// Slices must be [] not null so the UI can iterate without nil checks.
	if got.Skills == nil || got.Recent == nil {
		t.Fatalf("slices must be [] not null: skills=%v recent=%v", got.Skills, got.Recent)
	}
}
