// Package store is the ONLY package that contains SQL. It owns the SQLite
// database — a derived, rebuildable cache of the agent's own files. All writes
// funnel through here behind a single write mutex (WAL lets reads stay
// concurrent), so the concurrency story is trivial to reason about.
package store

import (
	"database/sql"
	"fmt"
	"sync"

	_ "modernc.org/sqlite" // registers the pure-Go "sqlite" driver
)

// Store wraps the *sql.DB plus the write mutex that serialises all writers.
type Store struct {
	db   *sql.DB
	wmu  sync.Mutex // held by every write path; readers don't take it
	path string
	// costFn prices a token bundle in USD. The store stays pricing-agnostic: the
	// services layer injects this once at startup (SetCostFn). When set, ApplyIngest
	// stamps est_cost_usd on each rollup row it folds, so the hot read/broadcast path
	// never has to recompute cost over the whole table. A nil costFn means "no
	// pricing wired" → ingest leaves est_cost_usd untouched (the pre-fix behaviour).
	costFn func(model string, in, out, cacheRead, cacheWrite int64) float64
}

// SetCostFn injects the pricing function used to stamp est_cost_usd on rollup
// rows at ingest time. Call once during wiring, before ingestion starts, so the
// store can keep cost current without the read path ever rewriting the table.
func (s *Store) SetCostFn(fn func(model string, in, out, cacheRead, cacheWrite int64) float64) {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	s.costFn = fn
}

// dsn builds the connection string with our locked PRAGMAs (spec §7.1).
func dsn(path string) string {
	return "file:" + path +
		"?_pragma=busy_timeout(5000)" +
		"&_pragma=journal_mode(WAL)" +
		"&_pragma=foreign_keys(ON)" +
		"&_pragma=synchronous(NORMAL)" +
		"&_txlock=immediate"
}

// Open opens (or creates) the database, applies PRAGMAs and all pending
// migrations, and returns a ready Store.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	// One writer at a time keeps WAL happy and matches our mutex model.
	db.SetMaxOpenConns(1)
	s := &Store{db: db, path: path}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// DB exposes the underlying handle for read queries in this package's files.
func (s *Store) DB() *sql.DB { return s.db }

// IntegrityOK runs PRAGMA integrity_check; ok is true only when SQLite reports
// the literal "ok". Used by the startup ladder to decide quarantine.
func (s *Store) IntegrityOK() (bool, error) {
	var res string
	if err := s.db.QueryRow("PRAGMA integrity_check").Scan(&res); err != nil {
		return false, err
	}
	return res == "ok", nil
}

// Close checkpoints the WAL (TRUNCATE) so the main db file is complete on disk,
// then closes the handle. Safe to call once during graceful shutdown.
func (s *Store) Close() error {
	// Best-effort checkpoint; a failure here must not block shutdown.
	_, _ = s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	return s.db.Close()
}
