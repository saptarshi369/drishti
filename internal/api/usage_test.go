package api

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/store"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "drishti.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return NewServer("test", st)
}

// TestHandleUsageShape verifies /api/usage returns a 7-day snapshot as JSON.
func TestHandleUsageShape(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest("GET", "/api/usage", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var snap model.UsageSnapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snap); err != nil {
		t.Fatal(err)
	}
	if snap.WindowDays != 7 || len(snap.Days) != 7 || !snap.Estimate {
		t.Fatalf("bad snapshot: %+v", snap)
	}
	// nil slices must serialise as [] not null.
	if snap.ByProject == nil || snap.ByModel == nil || snap.Heatmap == nil {
		t.Fatalf("slices must be [] not null: %+v", snap)
	}
}

// TestHandleQuotaGated verifies /api/quota returns a 200 snapshot with
// available=false when no samples exist (the gated state, NOT an error).
func TestHandleQuotaGated(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest("GET", "/api/quota", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var snap model.QuotaSnapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snap); err != nil {
		t.Fatal(err)
	}
	if snap.Available {
		t.Fatalf("want gated (available=false), got %+v", snap)
	}
}

// TestHandleQuotaSampleStores verifies a valid POST stores both windows (204) and
// that a subsequent GET /api/quota reports available=true.
func TestHandleQuotaSampleStores(t *testing.T) {
	srv := newTestServer(t)
	body := `{"agent":"claude","plan":"max","source":"statusline",
		"five_hour":{"used_percentage":41,"resets_at_ms":9},
		"seven_day":{"used_percentage":12,"resets_at_ms":7}}`
	req := httptest.NewRequest("POST", "/api/quota/sample", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Fatalf("status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}

	getReq := httptest.NewRequest("GET", "/api/quota", nil)
	getRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(getRec, getReq)
	var snap model.QuotaSnapshot
	_ = json.Unmarshal(getRec.Body.Bytes(), &snap)
	if !snap.Available || snap.FiveHour == nil || snap.FiveHour.UsedPercentage != 41 {
		t.Fatalf("sample not stored: %+v", snap)
	}
}

// TestHandleQuotaSampleRejectsEmpty verifies a body with no usable window returns
// a typed 400 (never a 500, never a stack trace).
func TestHandleQuotaSampleRejectsEmpty(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest("POST", "/api/quota/sample", bytes.NewBufferString(`{"agent":"claude"}`))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	var env map[string]map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env["error"]["code"] != "quota_sample_invalid" {
		t.Fatalf("error code = %v", env["error"]["code"])
	}
}

// TestHandleQuotaSampleRejectsUnknownAgent verifies that a well-formed body
// referencing an unknown agent (e.g. "codex") returns a typed 400, NOT 500.
// v1 is Claude-only; silently falling back to claude would mis-attribute data.
func TestHandleQuotaSampleRejectsUnknownAgent(t *testing.T) {
	srv := newTestServer(t)
	body := `{"agent":"codex","five_hour":{"used_percentage":1,"resets_at_ms":1}}`
	req := httptest.NewRequest("POST", "/api/quota/sample", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("unknown agent: status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	var env map[string]map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("response not a typed error envelope: %v; body=%s", err, rec.Body.String())
	}
	if env["error"]["code"] != "quota_sample_invalid" {
		t.Fatalf("error code = %v, want quota_sample_invalid", env["error"]["code"])
	}
}

// TestSnapshotMessagesIncludesQuota verifies the reconnect snapshot carries a
// "quota" frame so a freshly connected client heals its gauges without a refresh.
func TestSnapshotMessagesIncludesQuota(t *testing.T) {
	srv := newTestServer(t)
	var hasQuota bool
	for _, m := range srv.snapshotMessages() {
		if m.Type == "quota" {
			hasQuota = true
		}
	}
	if !hasQuota {
		t.Fatal("snapshotMessages missing quota frame")
	}
}
