// pkg/utils/math.go
package utils

import "math"

func RoundToCents(amount float64) float64 {
	return math.Round(amount*100) / 100
}

func Abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}