package services

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func qsnap(five, seven float64) model.QuotaSnapshot {
	return model.QuotaSnapshot{
		Available: true,
		FiveHour:  &model.QuotaWindow{UsedPercentage: five, TsMs: 100},
		SevenDay:  &model.QuotaWindow{UsedPercentage: seven, TsMs: 200},
	}
}

func TestBuildAlertsEmptyIsNonNil(t *testing.T) {
	got := BuildAlerts(AlertInputs{})
	if got == nil {
		t.Fatal("alerts must be non-nil (stable [] JSON)")
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestBuildAlertsQuotaThreshold(t *testing.T) {
	if a := BuildAlerts(AlertInputs{Quota: qsnap(79, 10)}); len(a) != 0 {
		t.Fatalf("79%% should not alert, got %d", len(a))
	}
	a := BuildAlerts(AlertInputs{Quota: qsnap(80, 10)})
	if len(a) != 1 || a[0].Kind != "quota" || a[0].Severity != "amber" {
		t.Fatalf("80%% want one amber quota alert, got %+v", a)
	}
	a = BuildAlerts(AlertInputs{Quota: qsnap(96, 10)})
	if len(a) != 1 || a[0].Severity != "red" {
		t.Fatalf("96%% want red, got %+v", a)
	}
	if a := BuildAlerts(AlertInputs{Quota: model.QuotaSnapshot{Available: false}}); len(a) != 0 {
		t.Fatalf("gated quota (no sample) should not alert, got %d", len(a))
	}
}

func TestBuildAlertsBlockedNamesOnly(t *testing.T) {
	a := BuildAlerts(AlertInputs{BlockedEvents: []model.RecentEvent{{TsMs: 5, ToolName: "Bash"}}})
	if len(a) != 1 || a[0].Kind != "blocked_command" || a[0].Severity != "red" {
		t.Fatalf("want one red blocked alert, got %+v", a)
	}
	if a[0].Text != "Command blocked — Bash" {
		t.Errorf("text = %q (must name the tool, never raw command text)", a[0].Text)
	}
}

func TestBuildAlertsSecurityAndDeadSkills(t *testing.T) {
	a := BuildAlerts(AlertInputs{SecurityHigh: 2, DeadSkills: 1})
	if len(a) != 2 {
		t.Fatalf("want 2 alerts, got %d", len(a))
	}
	if a[0].Kind != "security" || a[1].Kind != "dead_skill" {
		t.Errorf("order = %s,%s want security,dead_skill (red before grey)", a[0].Kind, a[1].Kind)
	}
	if a[0].Text != "2 high-severity security findings" {
		t.Errorf("sec text = %q", a[0].Text)
	}
	if a[1].Text != "1 dead skill (never fired)" {
		t.Errorf("dead text = %q", a[1].Text)
	}
}

func TestBuildAlertsSeverityOrdering(t *testing.T) {
	a := BuildAlerts(AlertInputs{
		Quota:         qsnap(85, 10),                                    // amber
		BlockedEvents: []model.RecentEvent{{TsMs: 9, ToolName: "Bash"}}, // red
		DeadSkills:    1,                                                // grey
	})
	if len(a) != 3 {
		t.Fatalf("want 3, got %d", len(a))
	}
	if a[0].Severity != "red" || a[1].Severity != "amber" || a[2].Severity != "grey" {
		t.Errorf("severity order = %s,%s,%s want red,amber,grey", a[0].Severity, a[1].Severity, a[2].Severity)
	}
}
