package main

import "math"

// Average calculates the arithmetic mean of a slice of floats.
func Average(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	var total float64
	for _, v := range values {
		total += v
	}
	return total / float64(len(values))
}

// Variance calculates the sample variance of a slice of floats.
func Variance(values []float64) float64 {
	if len(values) < 2 {
		return 0.0
	}
	avg := Average(values)
	var sumSq float64
	for _, v := range values {
		diff := v - avg
		sumSq += diff * diff
	}
	return sumSq / float64(len(values)-1)
}

// StdDev calculates the standard deviation.
func StdDev(values []float64) float64 {
	return math.Sqrt(Variance(values))
}

// MinMax returns the minimum and maximum values in a slice.
func MinMax(values []float64) (min, max float64) {
	if len(values) == 0 {
		return 0, 0
	}
	min, max = values[0], values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}
