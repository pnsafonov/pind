package math_utils

import (
	"math"
)

// Round2 - 77.1501111 -> 77.15
func Round2(x float64) float64 {
	return math.Round(x*100) / 100
}

// IntDivideCeil - round up division: 3/2=2, 5/2=3
func IntDivideCeil(x int, y int) int {
	x0 := float64(x)
	y0 := float64(y)
	result := x0 / y0
	result0 := math.Ceil(result)
	return int(result0)
}

func IntDivide2Ceil(x int) int {
	return IntDivideCeil(x, 2)
}
