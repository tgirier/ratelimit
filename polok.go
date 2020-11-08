// Package polok implements API testing tool
// It enables load testing with rate-limiting and burst functionalities
package polok

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

// MaxQPS is responsible for enforcing a global rate limiting
type MaxQPS struct {
	Rate float64 // query per second
}

// Consume consumes n tokens from the bucket channel at a given rate.
// A token represents a single request.
func (l *MaxQPS) Consume(n int, bucket <-chan struct{}) (total int) {

	tickInterval := time.Duration(1e9/l.Rate) * time.Nanosecond
	tick := time.Tick(tickInterval)

	counter := 0

	for i := 0; i < n; i++ {
		<-tick
		<-bucket
		counter++
	}

	return counter
}

// Request makes a given request if it is able to post a token.
// A custom http client can be provided, otherwise http default client will be used
func Request(req *http.Request, bucket chan<- struct{}, client *http.Client) (*http.Response, error) {
	if client == nil {
		client = http.DefaultClient
	}

	if req == nil {
		return nil, errors.New("missing request")
	}

	bucket <- struct{}{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// RequestWithLimit launches a given number of requests to a given URL at a given rate.
// A custom http client can be provided, otherwise http default client will be used.
func RequestWithLimit(req *http.Request, reqNumber int, rate float64, client *http.Client) (number int, finalRate float64, err error) {

	var n int
	var wg sync.WaitGroup
	var wgReq sync.WaitGroup

	m := MaxQPS{
		Rate: rate,
	}

	tokens := make(chan struct{})

	wg.Add(1)
	go func() {
		n = m.Consume(reqNumber, tokens)
		wg.Done()
	}()

	start := time.Now()

	for i := 0; i < reqNumber; i++ {
		wgReq.Add(1)
		go func(req *http.Request, client *http.Client) {
			_, _ = Request(req, tokens, client)
			wgReq.Done()
		}(req, client)
	}

	wgReq.Wait()

	stop := time.Now()
	duration := stop.Sub(start).Seconds()

	wg.Wait()

	r := float64(n) / duration

	return n, r, nil
}
