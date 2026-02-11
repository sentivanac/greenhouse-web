package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"greenhouse/internal/model"
	"greenhouse/internal/storage"
)

type API struct {
	store *storage.SQLite
}

func New(store *storage.SQLite) *API {
	return &API{store: store}
}

func (a *API) Latest(w http.ResponseWriter, r *http.Request) {
	m, err := a.store.GetLatest()
	if err != nil {
		http.Error(w, "no data", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func (a *API) Recent(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := 10
	if l := q.Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	data, err := a.store.GetRecent(limit)
	if err != nil {
		http.Error(w, "no data", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (a *API) Range(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	fromStr := q.Get("from")
	toStr := q.Get("to")
	if fromStr == "" || toStr == "" {
		http.Error(w, "missing from/to", http.StatusBadRequest)
		return
	}

	from, err1 := strconv.ParseInt(fromStr, 10, 64)
	to, err2 := strconv.ParseInt(toStr, 10, 64)
	if err1 != nil || err2 != nil || from >= to {
		http.Error(w, "invalid range", http.StatusBadRequest)
		return
	}

	data, step, err := a.store.GetRangeDownsampled(from, to)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Step int64               `json:"step"`
		Data []model.Measurement `json:"data"`
	}{
		Step: step,
		Data: data,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (a *API) RangeCombined(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	fromStr := q.Get("from")
	toStr := q.Get("to")
	if fromStr == "" || toStr == "" {
		http.Error(w, "missing from/to", http.StatusBadRequest)
		return
	}

	from, err1 := strconv.ParseInt(fromStr, 10, 64)
	to, err2 := strconv.ParseInt(toStr, 10, 64)
	if err1 != nil || err2 != nil || from >= to {
		http.Error(w, "invalid range", http.StatusBadRequest)
		return
	}

	env, step, err := a.store.GetRangeEnvelope(from, to)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}

	resp := RangeCombinedResponse{Step: step}

	for _, e := range env {
		resp.Data = append(resp.Data, RangeCombinedPoint{
			Ts: e.Ts,

			TempAvg: e.TempAvg,
			TempMin: e.TempMin,
			TempMax: e.TempMax,

			HumAvg: e.HumAvg,
			HumMin: e.HumMin,
			HumMax: e.HumMax,

			LightAvg: e.LightAvg,
			LightMin: e.LightMin,
			LightMax: e.LightMax,

			SoilAvg: e.SoilAvg,
			SoilMin: e.SoilMin,
			SoilMax: e.SoilMax,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
