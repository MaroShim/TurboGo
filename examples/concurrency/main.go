package main

import (
	"fmt"
	"sync"
	"time"
)

// Worker computes squares of numbers concurrently.
func Worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for n := range jobs {
		// Simulate computation time
		time.Sleep(20 * time.Millisecond)
		results <- n * n
	}
}

func main() {
	const numJobs = 8
	const numWorkers = 3

	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)
	var wg sync.WaitGroup

	fmt.Printf("--- Concurrent Worker Pool (%d workers, %d jobs) ---\n", numWorkers, numJobs)

	// Spawn workers
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go Worker(w, jobs, results, &wg)
	}

	// Send jobs
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	// Wait for workers to complete in background
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	total := 0
	for res := range results {
		fmt.Printf(" [Result] Computed square: %d\n", res)
		total += res
	}

	fmt.Printf("Done! Sum of squares: %d\n", total)
}
