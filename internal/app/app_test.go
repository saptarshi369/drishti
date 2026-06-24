package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saptarshi369/drishti/internal/config"
)

// TestInventoryLocations verifies that inventoryLocations always includes the
// user-global location and one extra entry per configured root. When no roots
// are configured it defaults to the user's home directory as the project root
// (M7 change: previously used cwd, but ~ is more useful for a permanent install).
func TestInventoryLocations(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// No roots: user-global PLUS the home directory as a default root.
	locs := inventoryLocations(config.Config{})
	if len(locs) != 2 {
		t.Fatalf("expected 2 locations with no roots (user-global + home), got %d", len(locs))
	}
	want := filepath.Join(home, ".claude")
	if locs[0].UserClaudeDir != want {
		t.Errorf("UserClaudeDir = %q; want %q", locs[0].UserClaudeDir, want)
	}
	if locs[0].ProjectRoot != "" {
		t.Errorf("user-global location must have empty ProjectRoot, got %q", locs[0].ProjectRoot)
	}
	if locs[1].ProjectRoot != home {
		t.Errorf("default project root = %q; want home %q", locs[1].ProjectRoot, home)
	}

	// Two roots: should return 1 + 2 = 3 locations.
	cfg := config.Config{Roots: []string{"/proj/a", "/proj/b"}}
	locs = inventoryLocations(cfg)
	if len(locs) != 3 {
		t.Fatalf("expected 3 locations with 2 roots, got %d", len(locs))
	}
	if locs[1].ProjectRoot != "/proj/a" {
		t.Errorf("locs[1].ProjectRoot = %q; want /proj/a", locs[1].ProjectRoot)
	}
	if locs[2].ProjectRoot != "/proj/b" {
		t.Errorf("locs[2].ProjectRoot = %q; want /proj/b", locs[2].ProjectRoot)
	}
}

func TestOpenStoreQuarantinesCorruptDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "drishti.db")
	// Write garbage so integrity_check fails (not a valid SQLite header).
	os.WriteFile(dbPath, []byte("this is not a sqlite database file at all"), 0o644)

	st, rebuilt, err := OpenStoreWithQuarantine(dbPath, nil)
	if err != nil {
		t.Fatalf("quarantine path must recover, got %v", err)
	}
	defer st.Close()
	if !rebuilt {
		t.Errorf("expected rebuilt=true for corrupt db")
	}
	matches, _ := filepath.Glob(dbPath + ".corrupt.*")
	if len(matches) == 0 {
		t.Errorf("expected a quarantined sidecar file")
	}
	ok, _ := st.IntegrityOK()
	if !ok {
		t.Errorf("rebuilt db should pass integrity_check")
	}
}

// TestEffectiveRootsDefaultsToHome verifies the M7 change: with no configured
// roots, the daemon watches the user's home dir (not the cwd), because pwd is
// meaningless once Drishti is installed somewhere permanent.
func TestEffectiveRootsDefaultsToHome(t *testing.T) {
	cfg := config.Default()
	got := effectiveRoots(cfg)
	home, _ := os.UserHomeDir()
	if len(got) != 1 || got[0] != home {
		t.Errorf("effectiveRoots = %v, want [%s]", got, home)
	}
}

func TestEffectiveRootsHonorsConfigured(t *testing.T) {
	cfg := config.Default()
	cfg.Roots = []string{"/tmp/proj"}
	if got := effectiveRoots(cfg); len(got) != 1 || got[0] != "/tmp/proj" {
		t.Errorf("effectiveRoots = %v, want [/tmp/proj]", got)
	}
}

func TestAcquireLockIsExclusive(t *testing.T) {
	dir := t.TempDir()
	release, err := acquireLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := acquireLock(dir); err == nil {
		t.Errorf("second lock must fail while first is held")
	}
	release()
	release2, err := acquireLock(dir)
	if err != nil {
		t.Errorf("lock should be reacquirable after release, got %v", err)
	}
	release2()
}
