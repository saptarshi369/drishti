package services

import (
	"math"
	"testing"
)

func TestCostKnownModel(t *testing.T) {
	// 1,000,000 input tokens at $15/M = $15.00 exactly.
	got := Cost("claude-opus-4-8", 1_000_000, 0, 0, 0)
	if math.Abs(got-15.0) > 1e-9 {
		t.Errorf("cost = %v, want 15.0", got)
	}
}

func TestCostUnknownModelIsZero(t *testing.T) {
	if got := Cost("mystery-model", 1_000_000, 1_000_000, 0, 0); got != 0 {
		t.Errorf("unknown model cost = %v, want 0", got)
	}
}
