package mqtt

import (
	"testing"
)

func TestParseMeasurement_OK(t *testing.T) {
	payload := []byte(`{
		"ts": 1770120576782,
		"t": 24.9,
		"rh": 23.12,
		"p": 1003.3,
		"light": 39
	}`)

	m, err := ParseMeasurement(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.T != 24.9 {
		t.Errorf("expected T=24.9, got %v", m.T)
	}

	if m.RH != 23.12 {
		t.Errorf("expected RH=23.12, got %v", m.RH)
	}
}

func TestParseMeasurement_InvalidJSON(t *testing.T) {
	payload := []byte(`{ this is not json }`)

	_, err := ParseMeasurement(payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
