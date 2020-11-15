// Package polok implements API testing tool
// It enables load testing with rate-limiting and burst functionalities
package polok

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

// Limit is responsible for enforcing a global rate limiting expressed in requests per seconds
// Limit consumes a given number of tokens from the bucket channel at a given rate.
// A token represents a single request.
func Limit(number int, rate float64, bucket <-chan struct{}) {

	tickInterval := time.Duration(1e9/rate) * time.Nanosecond
	tick := time.Tick(tickInterval)

	counter := 0

	for i := 0; i < number; i++ {
		<-tick
		<-bucket
		counter++
	}
}

// Request makes a given request if it is able to post a token.
// A custom http client can be provided, otherwise http default client will be used
func Request(req *http.Request, client *http.Client, bucket chan<- struct{}, reporting chan<- *http.Response) error {
	if client == nil {
		client = http.DefaultClient
	}

	if req == nil {
		return errors.New("missing request")
	}

	bucket <- struct{}{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	reporting <- resp

	return nil
}

// Report tracks the progress of given number of requests
// It aggregates the results within a list of responses
func Report(number int, reporting <-chan *http.Response) (responses []*http.Response) {
	var resp []*http.Response

	for i := 0; i < number; i++ {
		r := <-reporting
		resp = append(resp, r)
	}

	return resp
}

// RequestWithLimit launches a given number of requests to a given URL at a given rate.
// A custom http client can be provided, otherwise http default client will be used.
func RequestWithLimit(req *http.Request, reqNumber int, rate float64, client *http.Client) (responses []*http.Response, finalRate float64, err error) {

	var wg sync.WaitGroup
	var wgReq sync.WaitGroup

	tokens := make(chan struct{})
	reporting := make(chan *http.Response, reqNumber)

	wg.Add(1)
	go func() {
		Limit(reqNumber, rate, tokens)
		wg.Done()
	}()

	start := time.Now()

	for i := 0; i < reqNumber; i++ {
		wgReq.Add(1)
		go func(req *http.Request, client *http.Client) {
			_ = Request(req, client, tokens, reporting)
			wgReq.Done()
		}(req, client)
	}

	wgReq.Wait()

	stop := time.Now()
	duration := stop.Sub(start).Seconds()

	wg.Wait()

	resp := Report(reqNumber, reporting)

	r := float64(reqNumber) / duration

	return resp, r, nil
}
