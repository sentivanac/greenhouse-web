package mqtt

import (
	"encoding/json"

	"greenhouse/internal/model"
)

func ParseMeasurement(payload []byte) (model.Measurement, error) {
	var m model.Measurement

	if err := json.Unmarshal(payload, &m); err != nil {
		return m, err
	}
	return m, nil
}
