package app

import (
	"fmt"
	"io"
	"log/slog"
)

// safego runs fn in a goroutine, converting a panic into a logged error instead
// of a crashed process (spec §14: never panic in normal flow). In the thin slice
// it logs and returns; v1 hardens this into restart-with-backoff for long-lived
// loops. lg may be nil.
func safego(name string, lg *slog.Logger, fn func()) {
	if lg == nil {
		lg = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				lg.Error("recovered from panic", "goroutine", name, "panic", fmt.Sprint(r))
			}
		}()
		fn()
	}()
}
