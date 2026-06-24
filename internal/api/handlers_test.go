package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/saptarshi369/drishti/internal/store"
)

func TestHealthEndpoint(t *testing.T) {
	st, _ := store.Open(filepath.Join(t.TempDir(), "drishti.db"))
	defer st.Close()
	srv := NewServer("test", st)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string]any
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("status field = %v, want ok", body["status"])
	}
}

func TestOverviewEndpointShape(t *testing.T) {
	st, _ := store.Open(filepath.Join(t.TempDir(), "drishti.db"))
	defer st.Close()
	srv := NewServer("test", st)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/overview", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body)
	}
	var body struct {
		KPIs struct {
			PromptsToday  int     `json:"prompts_today"`
			SpendTodayUSD float64 `json:"spend_today_usd"`
		} `json:"kpis"`
		Recent []any `json:"recent"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("overview not valid JSON of expected shape: %v", err)
	}
}

func TestOverviewEndpointCarriesM8Fields(t *testing.T) {
	st, _ := store.Open(filepath.Join(t.TempDir(), "drishti.db"))
	defer st.Close()
	srv := NewServer("test", st)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/overview", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body)
	}
	var body struct {
		Health struct {
			Score int   `json:"score"`
			Bars  []any `json:"bars"`
		} `json:"health"`
		Alerts []any `json:"alerts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("overview not valid JSON of expected shape: %v", err)
	}
	if len(body.Health.Bars) != 4 {
		t.Errorf("health bars = %d, want 4", len(body.Health.Bars))
	}
	if body.Alerts == nil {
		t.Error("alerts must be non-null ([] not null)")
	}
}

func TestUpdateStatusEndpoint(t *testing.T) {
	srv := NewServer("v9.9.9", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/update/status", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	var body struct {
		Current   string `json:"current"`
		Available bool   `json:"available"`
	}
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Current != "v9.9.9" || body.Available {
		t.Errorf("update status = %+v, want current v9.9.9, available false", body)
	}
}
