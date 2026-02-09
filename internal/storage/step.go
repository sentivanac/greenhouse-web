package storage

const TargetPoints = 300

var allowedSteps = []int64{
	60_000,        // 1 min
	300_000,       // 5 min
	900_000,       // 15 min
	1_800_000,     // 30 min
	3_600_000,     // 1 h
	10_800_000,    // 3 h
	21_600_000,    // 6 h
	43_200_000,    // 12 h
	86_400_000,    // 1 day
}

func pickStep(from, to int64) int64 {
	if to <= from {
		return allowedSteps[0]
	}

	raw := (to - from) / TargetPoints
	for _, s := range allowedSteps {
		if s >= raw {
			return s
		}
	}
	return allowedSteps[len(allowedSteps)-1]
}
