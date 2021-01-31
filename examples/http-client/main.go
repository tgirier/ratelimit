package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/tgirier/ratelimit"
)

func main() {
	var wg sync.WaitGroup

	requests := []string{"https://www.google.fr", "https://www.github.com"} // Requests
	rate := 1.0                                                             // Rate: 1.0 query per second

	// Create a rate limited HTTP client
	c := ratelimit.NewHTTPClient(rate)
	fmt.Println("Rate limited HTTP client initialized")

	start := time.Now() // Start a timer to calculate the effective rate
	fmt.Printf("Starting to send request at a rate of %.2f QPS\n", rate)

	// Send requests concurrently
	wg.Add(len(requests))
	for _, req := range requests {
		go func(req string) {
			c.GetWithRateLimit(req)
			fmt.Printf("Request to %v sent\n", req)
			wg.Done()
		}(req)
	}

	wg.Wait()

	// Calculate the effective rate
	stop := time.Now()
	duration := stop.Sub(start).Seconds()
	effectiveRate := float64(len(requests)) / duration

	fmt.Printf("Sent %d requests at an effective rate of %.2f QPS\n", len(requests), effectiveRate)

}
