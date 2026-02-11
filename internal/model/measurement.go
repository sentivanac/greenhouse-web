package model

type Measurement struct {
	Ts    int64   `json:"ts"`
	T     float64 `json:"t"`
	RH    float64 `json:"rh"`
	P     float64 `json:"p"`
	Light int     `json:"light"`
	Soil  int     `json:"soil"`
}
