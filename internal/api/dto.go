package api

type RangeCombinedPoint struct {
	Ts int64 `json:"ts"`

	TempAvg float64 `json:"temp_avg"`
	TempMin float64 `json:"temp_min"`
	TempMax float64 `json:"temp_max"`

	HumAvg float64 `json:"hum_avg"`
	HumMin float64 `json:"hum_min"`
	HumMax float64 `json:"hum_max"`


	LightAvg float64 `json:"light_avg"`
	LightMin float64 `json:"light_min"`
	LightMax float64 `json:"light_max"`
}

type RangeCombinedResponse struct {
	Step int64                `json:"step"`
	Data []RangeCombinedPoint `json:"data"`
}
