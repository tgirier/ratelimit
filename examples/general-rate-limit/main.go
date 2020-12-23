package main

import (
	"fmt"
	"time"

	"github.com/tgirier/ratelimit"
)

func main() {
	rate := 1.0 // Rate: 1.0 executions per seconds
	n := 2      // Number of times the function will be executed

	// Create a rate limited worker
	w := ratelimit.NewWorker(rate, func() { fmt.Println("Hello World!") })
	fmt.Println("Rate limited worker initialized")

	start := time.Now() // Start a timer to calculate the effective rate
	fmt.Printf("Starting to execute the provided function at a rate of %.2f QPS\n", rate)

	// Execute function
	for i := 0; i < n; i++ {
		w.DoWithRateLimit()
	}

	// Calculate the effective rate
	stop := time.Now()
	duration := stop.Sub(start).Seconds()
	effectiveRate := float64(n) / duration

	fmt.Printf("Executed the function %d times at an effective rate of %.2f QPS\n", n, effectiveRate)
}
