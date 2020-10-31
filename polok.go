// Package polok implements API testing tool
// It enables load testing with rate-limiting and burst functionalities
package polok

import (
	"errors"
	"net/http"
	"time"
)

// MaxQPS is responsible for enforcing a global rate limiting
type MaxQPS struct {
	Rate float64 // query per second
}

// Consume consumes n tokens from the bucket channel at a given rate.
// A token represents a single request.
func (l *MaxQPS) Consume(n int, bucket <-chan struct{}) (total int, rate float64) {

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
type Worker struct {
	Client *http.Client
}

// Request makes a given request if it is able to post a token.
func (w *Worker) Request(req *http.Request, bucket chan<- struct{}) (*http.Response, error) {
	if w.Client == nil {
		w.Client = http.DefaultClient
	}

	if req == nil {
		return nil, errors.New("missing request")
	}

	bucket <- struct{}{}

	resp, err := w.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
