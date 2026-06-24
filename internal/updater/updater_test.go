package updater

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
)

func TestCurrentStatusReportsVersionNoNetwork(t *testing.T) {
	s := CurrentStatus("v1.2.3")
	if s.Current != "v1.2.3" {
		t.Errorf("current = %q, want v1.2.3", s.Current)
	}
	if s.Available {
		t.Errorf("slice stub must never claim an update is available")
	}
	if len(s.Commands) == 0 {
		t.Fatal("expected upgrade commands for the OS")
	}
	joined := strings.Join(s.Commands, " ")
	if runtime.GOOS == "windows" && !strings.Contains(joined, "drishti.exe") {
		t.Errorf("windows commands should reference drishti.exe; got %v", s.Commands)
	}
	if runtime.GOOS != "windows" && !strings.Contains(joined, "go build -o drishti ") {
		t.Errorf("unix commands should reference the plain build; got %v", s.Commands)
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		cur, latest string
		newer       bool
	}{
		{"v1.2.0", "v1.3.0", true},
		{"v1.3.0", "v1.3.0", false},
		{"v1.4.0", "v1.3.0", false},
		{"dev", "v1.0.0", false}, // dev builds never claim an update
		{"v1.2.0", "garbage", false},
	}
	for _, c := range cases {
		if got := compareVersions(c.cur, c.latest); got != c.newer {
			t.Errorf("compareVersions(%q,%q) = %v, want %v", c.cur, c.latest, got, c.newer)
		}
	}
}

func TestCheckReportsNewer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"tag_name":"v9.9.9"}`))
	}))
	defer srv.Close()
	st := checkAt(context.Background(), "v1.0.0", srv.Client(), srv.URL)
	if !st.Available || st.Latest != "v9.9.9" {
		t.Errorf("status = %+v, want available v9.9.9", st)
	}
}

// TestCheckDegradesOnError verifies the spec §14 guarantee: any network
// failure (here an unreachable address) must NOT claim an update is available.
// The current version must be preserved so the UI can still display it.
func TestCheckDegradesOnError(t *testing.T) {
	// 127.0.0.1:0 is a valid-looking URL but nothing listens there; the TCP
	// dial will fail immediately (connection refused), exercising the error path.
	st := checkAt(context.Background(), "v1.0.0", http.DefaultClient, "http://127.0.0.1:0/bad")
	if st.Available {
		t.Errorf("offline check must not report an update: %+v", st)
	}
	if st.Current != "v1.0.0" {
		t.Errorf("current should be preserved, got %q", st.Current)
	}
}
