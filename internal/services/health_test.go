package services

import (
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func TestBuildHealthSubScoresAndComposite(t *testing.T) {
	tests := []struct {
		name      string
		in        HealthInputs
		wantBars  [4]int // ctx, sec, hyg, hook (fixed order)
		wantScore int
	}{
		{
			name:      "all clean",
			in:        HealthInputs{SecurityCounts: map[string]int{}},
			wantBars:  [4]int{100, 100, 100, 100},
			wantScore: 100,
		},
		{
			name:      "one high finding penalised 25",
			in:        HealthInputs{SecurityCounts: map[string]int{"high": 1}},
			wantBars:  [4]int{100, 75, 100, 100},
			wantScore: 94, // (100+75+100+100+2)/4 = 377/4 = 94
		},
		{
			name:      "context tax 40 percent",
			in:        HealthInputs{ContextPct: 40, SecurityCounts: map[string]int{}},
			wantBars:  [4]int{60, 100, 100, 100},
			wantScore: 90,
		},
		{
			name:      "skill hygiene half dead",
			in:        HealthInputs{SecurityCounts: map[string]int{}, ActiveSkills: 4, SkillCounts: model.SkillCounts{Dead: 2}},
			wantBars:  [4]int{100, 100, 50, 100},
			wantScore: 88, // (100+100+50+100+2)/4 = 352/4 = 88
		},
		{
			name:      "hook errors with hooks configured",
			in:        HealthInputs{SecurityCounts: map[string]int{}, HookErrors: 2, HooksConfigured: true},
			wantBars:  [4]int{100, 100, 100, 70},
			wantScore: 93, // (100+100+100+70+2)/4 = 372/4 = 93
		},
		{
			name:      "hook errors ignored when no hooks configured",
			in:        HealthInputs{SecurityCounts: map[string]int{}, HookErrors: 9, HooksConfigured: false},
			wantBars:  [4]int{100, 100, 100, 100},
			wantScore: 100,
		},
		{
			name:      "security floored at zero",
			in:        HealthInputs{SecurityCounts: map[string]int{"high": 10}},
			wantBars:  [4]int{100, 0, 100, 100},
			wantScore: 75,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildHealth(tt.in)
			if len(got.Bars) != 4 {
				t.Fatalf("bars = %d, want 4", len(got.Bars))
			}
			for i, w := range tt.wantBars {
				if got.Bars[i].Score != w {
					t.Errorf("bar[%d] (%s) = %d, want %d", i, got.Bars[i].Label, got.Bars[i].Score, w)
				}
			}
			if got.Score != tt.wantScore {
				t.Errorf("score = %d, want %d", got.Score, tt.wantScore)
			}
		})
	}
}
