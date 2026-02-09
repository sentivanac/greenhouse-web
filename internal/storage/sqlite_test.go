package storage

import (
	"testing"
	"time"

	"greenhouse/internal/model"
)

func TestInsertMeasurement(t *testing.T) {
	store, err := NewSQLite(":memory:")
	if err != nil {
		t.Fatalf("failed to init sqlite: %v", err)
	}

	m := model.Measurement{
		Ts:    time.Now().UnixMilli(),
		T:     24.9,
		RH:    23.12,
		P:     1003.3,
		Light: 39,
	}

	if err := store.Insert(m); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	count, err := store.Count()
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}

}

func TestGetLatest(t *testing.T) {
	store, err := NewSQLite(":memory:")
	if err != nil {
		t.Fatalf("failed to init sqlite: %v", err)
	}

	m1 := model.Measurement{
		Ts:    1000,
		T:     20.0,
		RH:    40.0,
		P:     1000.0,
		Light: 10,
	}

	m2 := model.Measurement{
		Ts:    2000,
		T:     25.0,
		RH:    50.0,
		P:     1010.0,
		Light: 20,
	}

	if err := store.Insert(m1); err != nil {
		t.Fatal(err)
	}

	if err := store.Insert(m2); err != nil {
		t.Fatal(err)
	}

	latest, err := store.GetLatest()
	if err != nil {
		t.Fatal(err)
	}

	if latest.Ts != m2.Ts {
		t.Fatalf("expected latest ts=%d, got %d", m2.Ts, latest.Ts)
	}

	if latest.T != m2.T {
		t.Fatalf("expected T=%.1f, got %.1f", m2.T, latest.T)
	}
}

func TestGetRange(t *testing.T) {
	store, err := NewSQLite(":memory:")
	if err != nil {
		t.Fatal(err)
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

	res, err := store.GetRange(1500, 2500)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}

	if res[0].Ts != 2000 {
		t.Fatalf("expected ts=2000, got %d", res[0].Ts)
	}
}

func TestGetRangeDownsampled(t *testing.T) {
	store, err := NewSQLite(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// simulacija: merenje svake minute tokom 1 sata
	start := int64(0)
	for i := 0; i < 60; i++ {
		m := model.Measurement{
			Ts: start + int64(i)*60_000,
			T:  float64(i),
		}
		if err := store.Insert(m); err != nil {
			t.Fatal(err)
		}
	}

	res, step, err := store.GetRangeDownsampled(0, 60*60_000)
	if err != nil {
		t.Fatal(err)
	}

	if step < 60_000 {
		t.Fatalf("step too small: %d", step)
	}

	if len(res) == 0 {
		t.Fatal("expected some downsampled points")
	}

	if len(res) > TargetPoints {
		t.Fatalf("too many points: %d", len(res))
	}
}

func TestGetRangeEnvelope(t *testing.T) {
	store, err := NewSQLite(":memory:")
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

	env, _, err := store.GetRangeEnvelope(0, 10_000)
	if err != nil {
		t.Fatal(err)
	}

	if len(env) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(env))
	}

	if env[0].Min != 10 || env[0].Max != 30 {
		t.Fatalf("unexpected envelope: %+v", env[0])
	}
}
