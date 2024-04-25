package pkg

// getIdlePart - returns part of cpu's cores,
func getIdlePart(numaCount int) float64 {
	switch numaCount {
	case 1:
		return 1.0 / 4.0 // 25 %
	case 2:
		return 1.0 / 3.0 // 33 %
	default: // >= 3
		return 0.5 // 50 %
	}
}

func getIdleCoresCountDefault(numaCount int, coresCount int) int {
	coresPart := getIdlePart(numaCount)
	coresCount0 := float64(coresCount)
	count := coresPart * coresCount0
	count0 := count + 0.01
	return int(count0)
}
