package store

import (
	"path/filepath"
	"testing"
)

func TestOpenAppliesMigrationsAndWAL(t *testing.T) {
	p := filepath.Join(t.TempDir(), "drishti.db")
	s, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	var mode string
	if err := s.DB().QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %q, want wal", mode)
	}

	var n int
	if err := s.DB().QueryRow("SELECT count(*) FROM schema_migrations").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n < 1 {
		t.Errorf("expected at least one applied migration, got %d", n)
	}

	var agent string
	if err := s.DB().QueryRow("SELECT code FROM agents WHERE id=1").Scan(&agent); err != nil {
		t.Fatal(err)
	}
	if agent != "claude" {
		t.Errorf("agent 1 = %q, want claude", agent)
	}
}

func TestMigrationsAreIdempotent(t *testing.T) {
	p := filepath.Join(t.TempDir(), "drishti.db")
	s, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	s2, err := Open(p)
	if err != nil {
		t.Fatalf("re-open failed: %v", err)
	}
	defer s2.Close()
}

func TestIntegrityOK(t *testing.T) {
	p := filepath.Join(t.TempDir(), "drishti.db")
	s, _ := Open(p)
	defer s.Close()
	ok, err := s.IntegrityOK()
	if err != nil || !ok {
		t.Errorf("integrity = %v, %v; want true, nil", ok, err)
	}
}
