package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"greenhouse/internal/model"
	"greenhouse/internal/storage"
)

func TestRangeCombined_OK(t *testing.T) {
	store, err := storage.NewSQLite(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// isti bucket, različite vrednosti
	values := []float64{10, 20, 30}
	for i, v := range values {
		m := model.Measurement{
			Ts: int64(i) * 1000,
			T:  v,
		}
		if err := store.Insert(m); err != nil {
			t.Fatal(err)
		}
	}

	api := New(store)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/range/combined?from=0&to=10000",
		nil,
	)
	rec := httptest.NewRecorder()

	api.RangeCombined(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp RangeCombinedResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	// --- asertije koje imaju smisla ---
	if resp.Step <= 0 {
		t.Fatalf("expected step > 0, got %d", resp.Step)
	}

	if len(resp.Avg) != 1 {
		t.Fatalf("expected 1 avg point, got %d", len(resp.Avg))
	}

	if resp.Avg[0].Value != 20 {
		t.Fatalf("expected avg=20, got %.2f", resp.Avg[0].Value)
	}

	if resp.Min[0].Value != 10 || resp.Max[0].Value != 30 {
		t.Fatalf("unexpected envelope: min=%.2f max=%.2f",
			resp.Min[0].Value, resp.Max[0].Value)
	}
}
