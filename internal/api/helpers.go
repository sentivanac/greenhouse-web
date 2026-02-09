package api

import (
	"net/http"
	"strconv"
)

func DELparseRangeOr400(w http.ResponseWriter, r *http.Request) (int64, int64, bool) {
	q := r.URL.Query()

	fromStr := q.Get("from")
	toStr := q.Get("to")
	if fromStr == "" || toStr == "" {
		http.Error(w, "missing from/to", http.StatusBadRequest)
		return 0, 0, false
	}

	from, err1 := strconv.ParseInt(fromStr, 10, 64)
	to, err2 := strconv.ParseInt(toStr, 10, 64)

	if err1 != nil || err2 != nil || from >= to {
		http.Error(w, "invalid range", http.StatusBadRequest)
		return 0, 0, false
	}

	return from, to, true
}
