package main

import (
	"fmt"
	"math"
	"sort"
)

// DataSet represents a collection of numerical data points.
type DataSet struct {
	Name   string
	Values []float64
}

// NewDataSet creates a new DataSet instance.
func NewDataSet(name string, values []float64) *DataSet {
	return &DataSet{
		Name:   name,
		Values: values,
	}
}

// Mean computes the arithmetic average of the data set.
func (d *DataSet) Mean() float64 {
	if len(d.Values) == 0 {
		return 0.0
	}
	total := 0.0
	for _, v := range d.Values {
		total += v
	}
	return total / float64(len(d.Values))
}

// Median computes the middle value of the sorted data set.
func (d *DataSet) Median() float64 {
	n := len(d.Values)
	if n == 0 {
		return 0.0
	}
	sorted := make([]float64, n)
	copy(sorted, d.Values)
	sort.Float64s(sorted)

	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2.0
}

// StdDev computes the sample standard deviation.
func (d *DataSet) StdDev() float64 {
	if len(d.Values) < 2 {
		return 0.0
	}
	mean := d.Mean()
	sumSquares := 0.0
	for _, v := range d.Values {
		diff := v - mean
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(len(d.Values)-1))
}

func main() {
	scores := []float64{88.5, 92.0, 79.5, 95.0, 84.0, 91.5, 76.0, 89.0}
	ds := NewDataSet("Midterm Exam Scores", scores)

	fmt.Printf("--- Statistics Report: %s ---\n", ds.Name)
	fmt.Printf("Count:              %d\n", len(ds.Values))
	fmt.Printf("Arithmetic Mean:    %.2f\n", ds.Mean())
	fmt.Printf("Median Value:       %.2f\n", ds.Median())
	fmt.Printf("Standard Deviation: %.2f\n", ds.StdDev())
}
