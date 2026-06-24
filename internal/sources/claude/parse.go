// Package claude reads Claude Code's files into plain model rows.
//
// The transcript parser is a PURE function: bytes in, model rows out, no DB and
// no globals, so it is trivially table-testable against fixtures. It tolerates
// unknown/missing JSON fields (forward-compatible) and never assumes optional
// fields exist. It stores NO raw prompt text (privacy default D8).
package claude

import (
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/saptarshi369/drishti/internal/model"
)

// contentBlock is one element of an assistant/user message content array.
// Claude Code encodes user turns as either a plain string OR an array of these
// blocks. Tool results come back as "tool_result" blocks; genuine human text
// arrives as "text" blocks (or as a bare string for the common simple case).
type contentBlock struct {
	Type      string          `json:"type"`        // text|tool_use|tool_result|thinking|…
	ID        string          `json:"id"`          // for tool_use blocks
	Name      string          `json:"name"`        // for tool_use blocks
	Input     json.RawMessage `json:"input"`       // for tool_use blocks
	IsError   bool            `json:"is_error"`    // for tool_result blocks
	ToolUseID string          `json:"tool_use_id"` // for tool_result blocks
	Content   json.RawMessage `json:"content"`     // nested content (tool_result body)
	Text      string          `json:"text"`        // for text blocks
}

// ParseContext carries pure-function configuration. DayFn maps an event's epoch
// millis to the local yyyymmdd it belongs to; nil means "use the local zone".
type ParseContext struct {
	DayFn func(tsMs int64) int
}

// rawLine is the subset of a transcript JSON object we care about.
type rawLine struct {
	Type      string      `json:"type"`
	Timestamp string      `json:"timestamp"`
	SessionID string      `json:"sessionId"`
	IsMeta    bool        `json:"isMeta"`
	Message   *rawMessage `json:"message"`
}

// rawMessage is the subset of a transcript message object we care about.
// Content holds the raw JSON value of message.content, which may be a JSON
// string (plain user prompt) or a JSON array (structured block list). Using
// json.RawMessage defers decoding so we can inspect the shape at call time.
type rawMessage struct {
	Model   string          `json:"model"`
	Usage   *rawUsage       `json:"usage"`
	Content json.RawMessage `json:"content"` // string OR []block; shape decides prompt-ness
}

// blocks decodes message.content into its array form. ok=false means the
// content was a JSON string (a plain user prompt), not an array. An absent or
// empty content field is also treated as non-array (not a tool_result turn).
func (m *rawMessage) blocks() (bs []contentBlock, ok bool) {
	if len(m.Content) == 0 {
		return nil, false
	}
	if m.Content[0] == '[' {
		// Tolerate partial decode; unknown block types are kept with just their
		// "type" field populated — sufficient for the isPrompt check below.
		_ = json.Unmarshal(m.Content, &bs)
		return bs, true
	}
	return nil, false // string content (e.g. `"hello"`)
}

// isPrompt reports whether a user line contains a genuine human prompt:
//   - string content  → true (most common case)
//   - array with at least one "text" block → true
//   - array of only "tool_result" blocks   → false (automated feedback, not human input)
func (m *rawMessage) isPrompt() bool {
	bs, isArray := m.blocks()
	if !isArray {
		// Either a plain string or an absent content field. Both represent a
		// user turn, not a tool_result batch, so we count it as a prompt.
		return true
	}
	for _, b := range bs {
		if b.Type == "text" {
			return true
		}
	}
	return false
}

type rawUsage struct {
	InputTokens              int64 `json:"input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
}

// Parse reads NDJSON transcript content and returns the extracted rows. It
// advances BytesConsumed only past lines terminated by a newline, so a partial
// trailing line is left for the next read.
func Parse(r io.Reader, ctx ParseContext) (model.ParseResult, error) {
	dayFn := ctx.DayFn
	if dayFn == nil {
		dayFn = localDay
	}
	var res model.ParseResult
	deltas := map[string]*model.SessionDelta{}

	br := bufio.NewReader(r)
	var consumed int64
	for {
		line, err := br.ReadBytes('\n')
		if len(line) > 0 && line[len(line)-1] == '\n' {
			consumed += int64(len(line))
			res.LinesConsumed++
			processLine(line, dayFn, &res, deltas)
		}
		if err != nil {
			break
		}
	}
	res.BytesConsumed = consumed
	for _, d := range deltas {
		res.Deltas = append(res.Deltas, *d)
	}
	return res, nil
}

// processLine parses one complete line, appending events/deltas. A malformed
// line is skipped, never fatal.
func processLine(line []byte, dayFn func(int64) int, res *model.ParseResult, deltas map[string]*model.SessionDelta) {
	var rl rawLine
	if err := json.Unmarshal(line, &rl); err != nil {
		res.ErrorCount++ // skip the bad line, never abort the file
		return
	}
	tsMs := parseTS(rl.Timestamp)

	switch rl.Type {
	case "user":
		// Guard: skip meta lines and lines with no message payload entirely.
		if rl.IsMeta || rl.Message == nil {
			return
		}
		// Scan for tool_result blocks FIRST — even non-prompt user lines carry
		// them (automated feedback from tool calls). For each is_error block we
		// emit exactly one event: "blocked" when the text matches a denial
		// pattern, "error" for all other tool failures.
		// PRIVACY (D8): blockText returns the text transiently; it is fed only
		// into classifyBlocked and immediately discarded — never stored on Event
		// or anywhere persisted.
		if bs, isArray := rl.Message.blocks(); isArray {
			for _, b := range bs {
				if b.Type != "tool_result" || !b.IsError {
					continue
				}
				status := "error"
				if classifyBlocked(blockText(b.Content)) {
					status = "blocked"
				}
				res.Events = append(res.Events, model.Event{
					AgentCode: "claude", TypeCode: status, SourceCode: "transcript",
					SessionID: rl.SessionID, TsMs: tsMs, Status: status,
					DedupeKey: fmt.Sprintf("claude|%s|result|%s", rl.SessionID, b.ToolUseID),
				})
			}
		}
		// A user line is a prompt ONLY if it carries string content or a
		// content array that contains at least one "text" block. Lines whose
		// content array consists solely of "tool_result" blocks are automated
		// feedback from tool calls — they look like user turns in the NDJSON
		// but are not human-initiated prompts and must not be counted.
		if !rl.Message.isPrompt() {
			return
		}
		res.Events = append(res.Events, model.Event{
			AgentCode: "claude", TypeCode: "prompt", SourceCode: "transcript",
			SessionID: rl.SessionID, TsMs: tsMs,
			DedupeKey: dedupeKey(rl.SessionID, line),
		})
		d := ensure(deltas, rl.SessionID, dayFn(tsMs), tsMs)
		d.PromptCount++
	case "assistant":
		if rl.Message == nil {
			return
		}
		// Emit tool_use or skill events for each tool_use block in the content
		// array. String content (no blocks) is silently skipped; we only act on
		// array-form content. Skill blocks (name=="Skill") get a "skill" event
		// and are NOT also counted as a tool_use — the continue enforces that.
		if bs, isArray := rl.Message.blocks(); isArray {
			for _, b := range bs {
				if b.Type != "tool_use" {
					// Skip text, thinking, tool_result, and any future block types.
					continue
				}
				if b.Name == "Skill" {
					// Skill invocations: emit a "skill" event carrying only the
					// skill name (e.g. "superpowers:brainstorming"). Never store
					// input.command or argument text — privacy default D8.
					res.Events = append(res.Events, model.Event{
						AgentCode: "claude", TypeCode: "skill", SourceCode: "transcript",
						SessionID: rl.SessionID, TsMs: tsMs, SkillName: skillNameOf(b.Input),
						DedupeKey: fmt.Sprintf("claude|%s|%s", rl.SessionID, b.ID),
					})
					continue
				}
				// Regular tool call (Bash, Read, Edit, WebSearch, etc.).
				res.Events = append(res.Events, model.Event{
					AgentCode: "claude", TypeCode: "tool_use", SourceCode: "transcript",
					SessionID: rl.SessionID, TsMs: tsMs, ToolName: b.Name,
					DedupeKey: fmt.Sprintf("claude|%s|%s", rl.SessionID, b.ID),
				})
			}
		}
		if rl.Message.Usage == nil {
			return
		}
		u := rl.Message.Usage
		d := ensure(deltas, rl.SessionID, dayFn(tsMs), tsMs)
		if rl.Message.Model != "" {
			d.Model = rl.Message.Model
		}
		d.InputTokens += u.InputTokens
		d.OutputTokens += u.OutputTokens
		d.CacheTokens += u.CacheCreationInputTokens + u.CacheReadInputTokens
	default:
		// attachment/system/mode/file-history-snapshot/etc: ignored in the slice.
	}
}

// ensure returns the accumulating delta for a session, creating it on first use.
func ensure(m map[string]*model.SessionDelta, sid string, day int, tsMs int64) *model.SessionDelta {
	d, ok := m[sid]
	if !ok {
		d = &model.SessionDelta{SessionID: sid, Day: day, StartedMs: tsMs}
		m[sid] = d
	}
	if tsMs > 0 && (d.StartedMs == 0 || tsMs < d.StartedMs) {
		d.StartedMs = tsMs
	}
	if d.Day == 0 {
		d.Day = day
	}
	return d
}

// dedupeKey builds the UNIQUE key "claude|<session>|<sha1(line)>".
func dedupeKey(sessionID string, line []byte) string {
	sum := sha1.Sum(line)
	return fmt.Sprintf("claude|%s|%x", sessionID, sum)
}

// parseTS converts an ISO-8601 timestamp to epoch millis; 0 if absent/unparseable.
func parseTS(s string) int64 {
	if s == "" {
		return 0
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return 0
	}
	return t.UnixMilli()
}

// localDay returns the yyyymmdd of tsMs in the machine's local zone.
func localDay(tsMs int64) int {
	t := time.UnixMilli(tsMs).Local()
	return t.Year()*10000 + int(t.Month())*100 + t.Day()
}

// blockText returns the plain text of a tool_result content (string form, or the
// concatenated text blocks of the array form). Used ONLY to classify blocked vs
// error; the result is discarded, never persisted (privacy D8).
//
// tool_result content may arrive as:
//   - a JSON string: `"Exit code 1\nfile listing"`
//   - a JSON array of text blocks: `[{"type":"text","text":"..."}]`
//
// Any other shape (nil, empty, unrecognised) returns "".
func blockText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	if raw[0] == '"' {
		// String form — the most common shape for tool_result content.
		var s string
		_ = json.Unmarshal(raw, &s)
		return s
	}
	// Array form — concatenate every text block's .text field.
	var bs []contentBlock
	if err := json.Unmarshal(raw, &bs); err != nil {
		return ""
	}
	var sb strings.Builder
	for _, b := range bs {
		sb.WriteString(b.Text)
		sb.WriteByte('\n')
	}
	return sb.String()
}

// skillNameOf extracts input.skill from a Skill tool_use block (e.g.
// "superpowers:brainstorming"); "" if absent/unparseable.
// We tolerate a nil/empty input gracefully — json.Unmarshal returns an error
// and we discard it, leaving in.Skill as "".
func skillNameOf(input json.RawMessage) string {
	// in is an anonymous struct used only to pluck the one field we need.
	// json.Unmarshal ignores fields it doesn't recognise, so unknown keys
	// (like "command") are silently dropped — privacy default D8 is satisfied
	// without any explicit filtering.
	var in struct {
		Skill string `json:"skill"`
	}
	_ = json.Unmarshal(input, &in)
	return in.Skill
}
