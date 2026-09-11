package main

import "fmt"

// Fibonacci calculates the n-th Fibonacci number iteratively.
func Fibonacci(n int) uint64 {
	if n <= 1 {
		return uint64(n)
	}
	var a, b uint64 = 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

func main() {
	fmt.Println("--- Fibonacci Sequence Calculator (Turbo Go) ---")
	for i := 0; i <= 15; i++ {
		fib := Fibonacci(i)
		fmt.Printf("F(%2d) = %d\n", i, fib)
	}
	fmt.Println("Done!")
}
