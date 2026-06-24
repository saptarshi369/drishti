package ingest

import (
	"context"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watch keeps the store current. It prefers fsnotify (event-driven) and falls
// back to a periodic stat-poll if the watcher cannot be created on this
// platform. It returns only when ctx is cancelled. Errors are logged, never
// fatal — a watcher problem must not crash the daemon.
func (r *Reconciler) Watch(ctx context.Context, onChange func(int)) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		r.lg.Warn("fsnotify unavailable; using poll fallback", "err", err)
		r.pollLoop(ctx, 3*time.Second, onChange)
		return
	}
	defer func() { _ = w.Close() }()
	for _, root := range r.roots {
		if err := w.Add(root); err != nil {
			r.lg.Warn("watch add failed", "root", root, "err", err)
		}
	}

	const debounce = 350 * time.Millisecond
	var timer *time.Timer
	timerC := make(<-chan time.Time)

	// Safety poll. fsnotify is NOT recursive: watching the top-level root (e.g.
	// ~/.claude/projects) does NOT report writes to transcripts inside per-project
	// SUBDIRECTORIES — which is the normal case (an active session appends to
	// ~/.claude/projects/<encoded-project>/<session>.jsonl). Without this, live
	// activity silently stops updating after startup until the daemon restarts.
	// A periodic re-scan catches nested writes; scanAllCount is stat-cheap and only
	// re-reads files whose size/mtime/identity actually changed, so idle polling is
	// light. fsnotify still provides low-latency reaction to top-level changes.
	const safetyPoll = 2 * time.Second
	poll := time.NewTicker(safetyPoll)
	defer poll.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-poll.C:
			n := r.scanAllCount()
			if onChange != nil {
				onChange(n)
			}
		case ev, ok := <-w.Events:
			if !ok {
				return
			}
			if ev.Has(fsnotify.Write) || ev.Has(fsnotify.Create) {
				// (Re)arm the debounce timer to coalesce write bursts.
				if timer == nil {
					timer = time.NewTimer(debounce)
					timerC = timer.C
				} else {
					timer.Reset(debounce)
				}
			}
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			r.lg.Warn("watcher error", "err", err)
		case <-timerC:
			timer, timerC = nil, make(<-chan time.Time)
			n := r.scanAllCount()
			if onChange != nil {
				onChange(n)
			}
		}
	}
}

// pollLoop is the watcher fallback trigger: it reconciles all known + new files
// on a fixed interval. The fsnotify-driven Watch (Task 17) reuses scanAllCount;
// poll exists for platforms where fsnotify is unavailable or errors. It returns
// only when ctx is cancelled.
func (r *Reconciler) pollLoop(ctx context.Context, every time.Duration, onChange func(int)) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			n := r.scanAllCount()
			if onChange != nil {
				onChange(n)
			}
		}
	}
}

// scanAllCount reconciles all known + new files and returns total new inserts.
func (r *Reconciler) scanAllCount() int {
	total := 0
	for _, root := range r.roots {
		_ = walkJSONL(root, func(p string) {
			if n, err := r.ReconcileFile(p); err == nil {
				total += n
			}
		})
	}
	return total
}
