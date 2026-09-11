package main

import "fmt"

func main() {
	fmt.Println("==================================================")
	fmt.Println("   Turbo Go - Multi-File Calculation Suite Demo   ")
	fmt.Println("==================================================")

	// 1. Math Module Tests (from math.go)
	fmt.Println("[1] Math Utilities (math.go):")
	n := 7
	fact := Factorial(n)
	fmt.Printf("    Factorial(%d) = %d\n", n, fact)

	a, b := uint64(48), uint64(18)
	gcdVal := Gcd(a, b)
	fmt.Printf("    GCD(%d, %d)   = %d\n", a, b, gcdVal)

	primes := []uint64{7, 11, 15, 19, 21, 23}
	fmt.Print("    Primes check: ")
	for _, p := range primes {
		if IsPrime(p) {
			fmt.Printf("%d(Y) ", p)
		} else {
			fmt.Printf("%d(N) ", p)
		}
	}
	fmt.Println()

	base, exp := uint64(2), 10
	pVal := Power(base, exp)
	fmt.Printf("    Power(%d, %d)   = %d\n", base, exp, pVal)

	// 2. Stats Module Tests (from stats.go)
	fmt.Println("\n[2] Statistics Utilities (stats.go):")
	data := []float64{12.5, 18.2, 9.4, 24.1, 15.8, 30.0, 8.6}
	avg := Average(data)
	variance := Variance(data)
	stdDev := StdDev(data)
	minVal, maxVal := MinMax(data)

	fmt.Printf("    Dataset:  %v\n", data)
	fmt.Printf("    Average:  %.2f\n", avg)
	fmt.Printf("    Variance: %.2f\n", variance)
	fmt.Printf("    StdDev:   %.2f\n", stdDev)
	fmt.Printf("    Min / Max: %.2f / %.2f\n", minVal, maxVal)

	fmt.Println("\n[Success] All multi-file modules executed properly!")
}
