package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"greenhouse/internal/model"
	"greenhouse/internal/storage"
)

func TestLatest_OK(t *testing.T) {
	store, err := storage.NewSQLite(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	m := model.Measurement{
		Ts:    1234,
		T:     25.5,
		RH:    50.0,
		P:     1010.0,
		Light: 42,
	}

	if err := store.Insert(m); err != nil {
		t.Fatal(err)
	}

	api := New(store)

	req := httptest.NewRequest(http.MethodGet, "/api/latest", nil)
	rec := httptest.NewRecorder()

	api.Latest(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var got model.Measurement
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}

	if got.T != m.T {
		t.Fatalf("expected T=%.1f, got %.1f", m.T, got.T)
	}
}

func TestLatest_NoData(t *testing.T) {
	store, err := storage.NewSQLite(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	api := New(store)

	req := httptest.NewRequest(http.MethodGet, "/api/latest", nil)
	rec := httptest.NewRecorder()

	api.Latest(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestRange_OK(t *testing.T) {
	store, err := storage.NewSQLite(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	var resp struct {
		Step int64               `json:"step"`
		Data []model.Measurement `json:"data"`
	}

	data := []model.Measurement{
		{Ts: 1000, T: 20},
		{Ts: 2000, T: 21},
		{Ts: 3000, T: 22},
	}
	for _, m := range data {
		if err := store.Insert(m); err != nil {
			t.Fatal(err)
		}
	}

	api := New(store)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/range?from=1500&to=2500",
		nil,
	)
	rec := httptest.NewRecorder()

	api.Range(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	got := resp.Data

	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}

	if got[0].T != 21 {
		t.Fatalf("expected avg T=21, got %.2f", got[0].T)
	}

	if got[0].Ts%resp.Step != 0 {
		t.Fatalf("timestamp not aligned to step: %d", got[0].Ts)
	}
}

func TestRange_InvalidParams(t *testing.T) {
	store, _ := storage.NewSQLite(":memory:")
	api := New(store)

	req := httptest.NewRequest(http.MethodGet, "/api/range?from=abc&to=123", nil)
	rec := httptest.NewRecorder()

	api.Range(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRange_Downsampled(t *testing.T) {
	store, err := storage.NewSQLite(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 288; i++ {
		m := model.Measurement{
			Ts: int64(i) * 300_000,
			T:  20.0,
		}
		if err := store.Insert(m); err != nil {
			t.Fatal(err)
		}
	}

	api := New(store)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/range?from=0&to=86400000",
		nil,
	)
	rec := httptest.NewRecorder()

	api.Range(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Step int64               `json:"step"`
		Data []model.Measurement `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if len(resp.Data) > storage.TargetPoints {
		t.Fatalf("too many points returned: %d", len(resp.Data))
	}

	if resp.Step == 0 {
		t.Fatal("step not returned")
	}
}
