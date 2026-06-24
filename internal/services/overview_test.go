package services

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/store"
)

func today() int {
	n := time.Now()
	return n.Year()*10000 + int(n.Month())*100 + n.Day()
}

func TestOverviewSnapshotAssembles(t *testing.T) {
	st, _ := store.Open(filepath.Join(t.TempDir(), "drishti.db"))
	defer st.Close()
	sf, _ := st.UpsertSourceFile(model.SourceFile{AgentCode: "claude", Kind: "transcript", AbsPath: "/x.jsonl", State: "active"})
	st.ApplyIngest(store.IngestBatch{
		SourceFileID: sf,
		Deltas: []model.SessionDelta{{
			SessionID: "s1", Model: "claude-opus-4-8", Day: today(),
			InputTokens: 1_000_000, OutputTokens: 0, PromptCount: 1, StartedMs: 1,
		}},
		Events:    []model.Event{{AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript", SessionID: "s1", TsMs: 1, DedupeKey: "claude|s1|x"}},
		NewOffset: 1,
	})

	ov, err := OverviewSnapshot(st, OverviewParams{Root: "", WindowTokens: 200_000})
	if err != nil {
		t.Fatal(err)
	}
	// Baseline KPIs + ticker still work (the slice guarantees).
	if ov.KPIs.PromptsToday != 1 {
		t.Errorf("prompts = %d, want 1", ov.KPIs.PromptsToday)
	}
	if ov.KPIs.SpendTodayUSD < 14.99 || ov.KPIs.SpendTodayUSD > 15.01 {
		t.Errorf("spend = %v, want ~15", ov.KPIs.SpendTodayUSD)
	}
	if len(ov.Recent) != 1 {
		t.Errorf("recent = %d, want 1", len(ov.Recent))
	}
	// New sections present with stable empty/zero shapes (no inventory loaded).
	if ov.ActiveComponents.Total != 0 {
		t.Errorf("active total = %d, want 0", ov.ActiveComponents.Total)
	}
	if len(ov.Health.Bars) != 4 {
		t.Fatalf("health bars = %d, want 4", len(ov.Health.Bars))
	}
	if ov.Health.Score != 100 {
		t.Errorf("health score = %d, want 100 (all clean)", ov.Health.Score)
	}
	if ov.Alerts == nil {
		t.Error("alerts must be non-nil (stable [] JSON)")
	}
	if len(ov.Alerts) != 0 {
		t.Errorf("alerts = %d, want 0", len(ov.Alerts))
	}
}
