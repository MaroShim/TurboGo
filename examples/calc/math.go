package main

// Factorial calculates n! iteratively.
func Factorial(n int) uint64 {
	if n <= 1 {
		return 1
	}
	var res uint64 = 1
	for i := 2; i <= n; i++ {
		res *= uint64(i)
	}
	return res
}

// Gcd calculates the greatest common divisor using Euclidean algorithm.
func Gcd(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// IsPrime checks if a positive integer is a prime number.
func IsPrime(n uint64) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	for i := uint64(5); i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

// Power calculates base^exp.
func Power(base uint64, exp int) uint64 {
	var result uint64 = 1
	b := base
	e := exp
	for e > 0 {
		if e%2 == 1 {
			result *= b
		}
		b *= b
		e /= 2
	}
	return result
}
