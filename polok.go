// Package polok implements API testing tool
// It enables load testing with rate-limiting and burst functionalities
package polok

import (
	"net/http"
	"time"
)

// Limiter is responsible for enforcing a global rate limiting
type Limiter struct {
	Rate float64 // query per second
}

// Consume consumes n tokens from the bucket channel at a given rate.
// A token represents a single request.
func (l *Limiter) Consume(n int, bucket <-chan struct{}) (total int, rate float64) {

	tickInterval := time.Duration(1e9/l.Rate) * time.Nanosecond
	tick := time.Tick(tickInterval)

	counter := 0
	start := time.Now()

	for i := 0; i < n; i++ {
		<-tick
		<-bucket
		counter++
	}

	stop := time.Now()
	duration := stop.Sub(start).Seconds()

	return counter, float64(counter) / duration
}

// Worker is responsible for requesting a given URL with a given method
type Worker struct{}

// Request makes a request to a given URL with a given method
func (w *Worker) Request(method string, url string) (int, error) {
	c := http.Client{}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return 0, err
	}

	sc := resp.StatusCode

	return sc, nil
}
