package api

import "testing"

func TestHubBroadcastReachesSubscriber(t *testing.T) {
	h := NewHub()
	ch, cancel := h.Subscribe()
	defer cancel()

	h.Broadcast(Message{Type: "counters", TS: 1, Payload: map[string]int{"n": 5}})

	got := <-ch
	if got.Type != "counters" {
		t.Fatalf("got type %q, want counters", got.Type)
	}
}
