package model

type Envelope struct {
	Ts int64

	TempAvg float64
	TempMin float64
	TempMax float64

	HumAvg float64
	HumMin float64
	HumMax float64
}
