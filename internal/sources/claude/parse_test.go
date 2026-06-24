package claude

import (
	"os"
	"strings"
	"testing"
)

func fixedDay(int64) int { return 20260621 }

func TestParseBasicCountsPromptsAndTokens(t *testing.T) {
	b, _ := os.ReadFile("../../../testdata/transcripts/basic.jsonl")
	res, err := Parse(strings.NewReader(string(b)), ParseContext{DayFn: fixedDay})
	if err != nil {
		t.Fatal(err)
	}
	if got := len(res.Events); got != 1 {
		t.Fatalf("events = %d, want 1 (the prompt)", got)
	}
	if res.Events[0].TypeCode != "prompt" || res.Events[0].SessionID != "s1" {
		t.Errorf("bad event: %+v", res.Events[0])
	}
	var in, out, cache int64
	var prompts int
	for _, d := range res.Deltas {
		in += d.InputTokens
		out += d.OutputTokens
		cache += d.CacheTokens
		prompts += d.PromptCount
	}
	if in != 6376 || out != 401 || cache != 2934+8139 {
		t.Errorf("tokens in=%d out=%d cache=%d", in, out, cache)
	}
	if prompts != 1 {
		t.Errorf("prompts = %d, want 1", prompts)
	}
	if res.BytesConsumed != int64(len(b)) {
		t.Errorf("consumed = %d, want %d (all complete lines)", res.BytesConsumed, len(b))
	}
}

func TestParseSkipsMalformedLineNoAbort(t *testing.T) {
	b, _ := os.ReadFile("../../../testdata/transcripts/malformed.jsonl")
	res, err := Parse(strings.NewReader(string(b)), ParseContext{DayFn: fixedDay})
	if err != nil {
		t.Fatalf("malformed line must not error: %v", err)
	}
	if len(res.Events) != 2 {
		t.Errorf("events = %d, want 2 good lines", len(res.Events))
	}
	if res.ErrorCount != 1 {
		t.Errorf("error_count = %d, want 1", res.ErrorCount)
	}
}

func TestParseDoesNotConsumePartialLastLine(t *testing.T) {
	b, _ := os.ReadFile("../../../testdata/transcripts/partial.jsonl")
	res, err := Parse(strings.NewReader(string(b)), ParseContext{DayFn: fixedDay})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Events) != 1 {
		t.Errorf("events = %d, want 1 (only the complete line)", len(res.Events))
	}
	firstLineLen := strings.Index(string(b), "\n") + 1
	if int(res.BytesConsumed) != firstLineLen {
		t.Errorf("consumed = %d, want %d (up to last newline)", res.BytesConsumed, firstLineLen)
	}
}

func TestParse_ToolUseAndSkillEvents(t *testing.T) {
	line := `{"type":"assistant","sessionId":"s1","timestamp":"2026-06-21T10:00:00Z","message":{"content":[` +
		`{"type":"tool_use","id":"tu1","name":"Bash","input":{"command":"ls"}},` +
		`{"type":"tool_use","id":"tu2","name":"Skill","input":{"skill":"superpowers:brainstorming"}}]}}` + "\n"
	res, err := Parse(strings.NewReader(line), ParseContext{})
	if err != nil {
		t.Fatal(err)
	}
	var tools, skills int
	var skillName, toolName, toolKey string
	for _, e := range res.Events {
		switch e.TypeCode {
		case "tool_use":
			tools++
			toolName = e.ToolName
			toolKey = e.DedupeKey
		case "skill":
			skills++
			skillName = e.SkillName
		}
	}
	if tools != 1 || toolName != "Bash" {
		t.Fatalf("want 1 tool_use Bash, got %d %q", tools, toolName)
	}
	if skills != 1 || skillName != "superpowers:brainstorming" {
		t.Fatalf("want 1 skill brainstorming, got %d %q", skills, skillName)
	}
	if toolKey != "claude|s1|tu1" {
		t.Fatalf("dedupe key = %q", toolKey)
	}
}

func TestParse_ErrorAndBlockedEvents(t *testing.T) {
	lines := `{"type":"user","sessionId":"s1","timestamp":"2026-06-21T10:00:00Z","message":{"content":[{"type":"tool_result","tool_use_id":"t1","is_error":true,"content":"Exit code 1\nfile listing"}]}}
{"type":"user","sessionId":"s1","timestamp":"2026-06-21T10:00:01Z","message":{"content":[{"type":"tool_result","tool_use_id":"t2","is_error":true,"content":"Permission to use Bash has been denied"}]}}
`
	res, _ := Parse(strings.NewReader(lines), ParseContext{})
	var errs, blocked int
	var blockedKey string
	for _, e := range res.Events {
		switch e.Status {
		case "error":
			errs++
		case "blocked":
			blocked++
			blockedKey = e.DedupeKey
		}
	}
	if errs != 1 || blocked != 1 {
		t.Fatalf("want 1 error + 1 blocked, got %d / %d", errs, blocked)
	}
	if blockedKey != "claude|s1|result|t2" {
		t.Fatalf("blocked dedupe key = %q", blockedKey)
	}
}

func TestParse_ToolResultUserLinesAreNotPrompts(t *testing.T) {
	// One genuine prompt (string content) + two tool_result user lines.
	lines := `{"type":"user","sessionId":"s1","timestamp":"2026-06-21T10:00:00Z","message":{"content":"hello"}}
{"type":"user","sessionId":"s1","timestamp":"2026-06-21T10:00:01Z","message":{"content":[{"type":"tool_result","tool_use_id":"t1","content":"ok"}]}}
{"type":"user","sessionId":"s1","timestamp":"2026-06-21T10:00:02Z","message":{"content":[{"type":"tool_result","tool_use_id":"t2","content":"ok"}]}}
`
	res, err := Parse(strings.NewReader(lines), ParseContext{})
	if err != nil {
		t.Fatal(err)
	}
	prompts := 0
	for _, e := range res.Events {
		if e.TypeCode == "prompt" {
			prompts++
		}
	}
	if prompts != 1 {
		t.Fatalf("want 1 prompt, got %d", prompts)
	}
}
