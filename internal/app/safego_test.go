package app

import (
	"testing"
	"time"
)

func TestSafegoRunsAndRecoversPanic(t *testing.T) {
	done := make(chan struct{})
	safego("ok", nil, func() { close(done) })
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("safego did not run the function")
	}

	// A panicking goroutine must be recovered, not crash the test process.
	recovered := make(chan struct{})
	safego("boom", nil, func() {
		defer close(recovered)
		panic("kaboom")
	})
	select {
	case <-recovered:
	case <-time.After(time.Second):
		t.Fatal("safego did not run the panicking function")
	}
}
