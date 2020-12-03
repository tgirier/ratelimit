package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/tgirier/ratelimit"
)

func main() {
	var wg sync.WaitGroup

	rate := 1.0                                 // Rate: 1.0 executions per seconds
	n := 2                                      // Number of times the function will be executed
	f := func() { fmt.Println("Hello World!") } // Function that will be executed

	// Create a rate limited worker
	w := ratelimit.NewWorker(rate, f)
	fmt.Println("Rate limited worker initialized")

	start := time.Now() // Start a timer to calculate the effective rate
	fmt.Printf("Starting to execute the provided function at a rate of %.2f QPS\n", rate)

	// Execute function concurrently
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			w.DoWithRateLimit()
			wg.Done()
		}()
	}

	wg.Wait()

	// Calculate the effective rate
	stop := time.Now()
	duration := stop.Sub(start).Seconds()
	effectiveRate := float64(n) / duration

	fmt.Printf("Executed the function %d times at an effective rate of %.2f QPS\n", n, effectiveRate)
}
