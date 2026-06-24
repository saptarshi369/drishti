package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/saptarshi369/drishti/internal/services"
)

// Message is the SSE envelope broadcast to every connected UI client.
type Message struct {
	// Type identifies the payload shape. Valid values:
	//   "counters"  — overall Snapshot (totals, session counts, etc.)
	//   "activity"  — ActivitySnapshot (per-agent skill/tool/token activity)
	//   "quota"     — QuotaSnapshot (live plan-quota gauges)
	//   "status"    — daemon liveness ping ({"state":"live"})
	Type    string `json:"type"`
	TS      int64  `json:"ts"`      // epoch-millis UTC
	Payload any    `json:"payload"` // shape depends on Type
}

// Hub fans one published Message out to every subscribed SSE connection.
// It is safe for concurrent use. Slow clients are dropped rather than allowed
// to block the publisher (backpressure: we favour the daemon's liveness).
type Hub struct {
	mu   sync.Mutex
	subs map[chan Message]struct{}
}

// NewHub constructs an empty Hub.
func NewHub() *Hub {
	return &Hub{subs: make(map[chan Message]struct{})}
}

// Subscribe registers a new client. It returns a receive-only channel and a
// cancel func the caller MUST invoke (e.g. on client disconnect) to release it.
func (h *Hub) Subscribe() (<-chan Message, func()) {
	ch := make(chan Message, 16) // small buffer absorbs brief bursts
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	cancel := func() {
		h.mu.Lock()
		if _, ok := h.subs[ch]; ok {
			delete(h.subs, ch)
			close(ch)
		}
		h.mu.Unlock()
	}
	return ch, cancel
}

// Broadcast delivers m to every subscriber without blocking: if a client's
// buffer is full it is skipped for this message (the next snapshot heals it).
func (h *Hub) Broadcast(m Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs {
		select {
		case ch <- m:
		default:
		}
	}
}

// streamHandler streams live updates to one client as text/event-stream. On
// connect it immediately writes a fresh snapshot so a reconnecting UI heals
// without a page refresh, then relays every Message until the client leaves.
func (s *Server) streamHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, cancel := s.hub.Subscribe()
	defer cancel()

	// Initial snapshot (counters + activity + quota + status) before any broadcast.
	for _, m := range s.snapshotMessages() {
		writeSSE(w, flusher, m)
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case m, ok := <-ch:
			if !ok {
				return
			}
			writeSSE(w, flusher, m)
		}
	}
}

// writeSSE encodes one Message as an SSE "data:" frame and flushes it.
func writeSSE(w http.ResponseWriter, flusher http.Flusher, m Message) {
	b, _ := json.Marshal(m)
	if _, err := w.Write([]byte("data: ")); err != nil {
		return
	}
	_, _ = w.Write(b)
	_, _ = w.Write([]byte("\n\n"))
	flusher.Flush()
}

// snapshotMessages builds the counters+activity+quota+status frames from the
// current store state. A nil store (or overview snapshot error) still yields a
// status frame so the client at least learns the daemon is live.
//
// Message order: counters → activity → quota → status. This order is load-bearing:
// the UI processes counters first (fast global totals), then activity (per-agent
// detail), then quota (plan gauges), then status (liveness confirmation). Do not
// reorder.
//
// Graceful degradation: if ActivitySnapshot fails (e.g. store query error), the
// activity message is simply omitted — counters and status are still sent. The
// daemon must never drop the whole snapshot just because activity is unavailable.
func (s *Server) snapshotMessages() []Message {
	now := time.Now().UnixMilli()
	// status is always included last so the client always gets a liveness signal.
	status := Message{Type: "status", TS: now, Payload: map[string]string{"state": "live"}}
	if s.st == nil {
		return []Message{status}
	}
	ov, err := services.OverviewSnapshot(s.st, s.overviewParams())
	if err != nil {
		// Overview failed; still return status so the UI knows the daemon is up.
		return []Message{status}
	}

	// Start with the overview counters message (always present when store is ok).
	msgs := []Message{{Type: "counters", TS: now, Payload: ov}}

	// Append activity snapshot if available. Failure is non-fatal: the activity
	// frame is omitted but counters+status are still delivered (§14 failsafe).
	if act, err := services.ActivitySnapshot(s.st, s.currentDefaultRoot()); err == nil {
		msgs = append(msgs, Message{Type: "activity", TS: now, Payload: act})
	}

	// Append quota snapshot if available. Failure is non-fatal — the gauges simply
	// stay in their last/gated state until the next sample arrives (§14 failsafe).
	if q, err := services.QuotaSnapshot(s.st, "claude"); err == nil {
		msgs = append(msgs, Message{Type: "quota", TS: now, Payload: q})
	}

	// status always trails the data messages.
	msgs = append(msgs, status)
	return msgs
}

// BroadcastSnapshot pushes a fresh counters+activity+quota+status snapshot to
// every connected client. The daemon calls this after ingest detects new data and on a timer.
func (s *Server) BroadcastSnapshot() {
	for _, m := range s.snapshotMessages() {
		s.hub.Broadcast(m)
	}
}
