package store

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// migrate applies every embedded migration whose version is not yet recorded,
// each in its own transaction, forward-only. Because the DB is a rebuildable
// cache there are no down-migrations: a hard failure escalates to quarantine
// (handled by the app layer), not a rollback.
func (s *Store) migrate() error {
	if _, err := s.db.Exec(
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY, name TEXT, checksum TEXT, applied_ms INTEGER)`,
	); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names) // NNNN_ prefix makes lexical order = apply order

	for _, name := range names {
		version, err := strconv.Atoi(strings.SplitN(name, "_", 2)[0])
		if err != nil {
			return fmt.Errorf("bad migration name %q: %w", name, err)
		}
		var exists int
		if err := s.db.QueryRow(
			"SELECT count(*) FROM schema_migrations WHERE version=?", version,
		).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}
		body, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		if err := s.applyOne(version, name, string(body)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
	}
	return nil
}

// applyOne runs one migration body plus its bookkeeping row in a single txn.
func (s *Store) applyOne(version int, name, body string) error {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(body); err != nil {
		return err
	}
	if _, err := tx.Exec(
		"INSERT INTO schema_migrations(version,name,checksum,applied_ms) VALUES(?,?,?,?)",
		version, name, "", time.Now().UnixMilli(),
	); err != nil {
		return err
	}
	return tx.Commit()
}
