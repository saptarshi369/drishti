package services

import (
	"testing"
	"time"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/store"
)

// TestActivitySnapshotRecentNeverNil verifies that a scope with no matching events
// still yields a non-nil (empty) Recent slice. The SSE activity frame serializes
// the snapshot directly, and the Overview does recent.slice(...) — a JSON null
// (from a Go nil slice) would crash the page. This became reachable once the live
// feed could be filtered to a folder that has no recent activity.
func TestActivitySnapshotRecentNeverNil(t *testing.T) {
	st := openStore(t)
	snap, err := ActivitySnapshot(st, "", "-no-such-project")
	if err != nil {
		t.Fatal(err)
	}
	if snap.Recent == nil {
		t.Error("ActivitySnapshot.Recent is nil; want a non-nil empty slice ([] not null)")
	}
	if snap.Skills == nil {
		t.Error("ActivitySnapshot.Skills is nil; want a non-nil empty slice ([] not null)")
	}
}

// TestActivitySnapshotAssembles inserts events with near-now timestamps (so
// they fall inside the 24-hour rolling window), calls ActivitySnapshot, and
// verifies the assembled payload has populated counters and sparkline slices.
//
// Why near-now timestamps? ActivitySnapshot calls st.ActivityCounters(dayAgo,"")
// where dayAgo = now - 24h. The brief's skeleton used TsMs:1000 (epoch+1s, far
// in the past), which would place the events OUTSIDE the 24h window and make
// Last24h.Prompts == 0 — a trivially passing but meaningless assertion.
// We use time.Now().UnixMilli() so the assertion snap.Counters.Last24h.Prompts >= 1
// is both correct and meaningful.
func TestActivitySnapshotAssembles(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/t.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })

	// Use current time so events land inside the 24h window.
	nowMs := time.Now().UnixMilli()
	if _, err := st.ApplyIngest(store.IngestBatch{Events: []model.Event{
		{AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript", SessionID: "s1", TsMs: nowMs, DedupeKey: "p1"},
		{AgentCode: "claude", TypeCode: "skill", SourceCode: "transcript", SessionID: "s1", TsMs: nowMs + 1, SkillName: "deploy", DedupeKey: "k1"},
	}}); err != nil {
		t.Fatal(err)
	}

	snap, err := ActivitySnapshot(st, "", "")
	if err != nil {
		t.Fatal(err)
	}
	// Last24h counters must reflect the inserted events.
	if snap.Counters.Last24h.Prompts < 1 || snap.Counters.Last24h.Skills < 1 {
		t.Fatalf("counters not populated: %+v", snap.Counters)
	}
	// Sparkline slice length is always sparkBuckets (30); never empty.
	if len(snap.Sparklines.PromptsPerMin) == 0 {
		t.Fatalf("sparklines empty")
	}
}
