package utils

import "math"

func f(x, a float64) float64 {

	return math.Exp(x) - a
}

func Ln(n float64) float64 {

	var lo, hi, m float64

	if n <= 0 {

		return -1
	}

	if n == 1 {

		return 0
	}

	EPS := 0.00001

	lo = 0

	hi = n

	for math.Abs(lo-hi) >= EPS {

		m = float64((lo + hi) / 2.0)

		if f(m, n) < 0 {

			lo = m

		} else {

			hi = m
		}
	}

	return float64((lo + hi) / 2.0)
}

func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func MaxInt(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func MinInt64(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func Sigmoid(x float64) float64 {
	return 1.0 / (1 + math.Exp(-x))
}
