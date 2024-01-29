package math_utils

import "math"

// Round2 - 77.1501111 -> 77.15
func Round2(x float64) float64 {
	return math.Round(x*100) / 100
}
