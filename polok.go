// Package polok implements API testing tool
// It enables load testing with rate-limiting and burst functionalities
package polok

import (
	"errors"
	"fmt"
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
func Request(req *http.Request, client *http.Client, bucket chan<- struct{}, reporting chan<- *http.Response, wg *sync.WaitGroup) error {
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

	wg.Done()

	return nil
}

// NewRequester makes HTTP requests.
// It listens to an incoming stream of requests and return a channel of corresponding responses.
func NewRequester(done <-chan struct{}, inputStream <-chan *http.Request, bucket chan<- struct{}, client *http.Client) (responseStream <-chan *http.Response) {
	resultStream := make(chan *http.Response)

	if client == nil {
		client = http.DefaultClient
	}

	go func() {
		defer close(resultStream)
		for {
			select {
			case <-done:
				return
			case req := <-inputStream:
				if req == nil {
					fmt.Println("missing request")
					continue
				}
				bucket <- struct{}{}
				res, err := client.Do(req)
				if err != nil {
					fmt.Printf("request failed, %v", err)
				}
				resultStream <- res
			}
		}
	}()
	return resultStream
}

// Requests makes a given list of requests.
// It launches as many Request goroutines as they are requests provided.
// It calculates the overall rate of the requests.
func Requests(requests []*http.Request, client *http.Client, bucket chan<- struct{}, reporting chan<- *http.Response) (rate float64, err error) {
	var wg sync.WaitGroup

	wg.Add(len(requests))

	start := time.Now()

	for _, req := range requests {
		go Request(req, client, bucket, reporting, &wg)
	}

	wg.Wait()

	stop := time.Now()
	duration := stop.Sub(start).Seconds()

	rate = float64(len(requests)) / duration

	return rate, nil
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
func RequestWithLimit(requests []*http.Request, rate float64, client *http.Client) (responses []*http.Response, finalRate float64, err error) {

	tokens := make(chan struct{})
	reporting := make(chan *http.Response, len(requests))

	go Limit(len(requests), rate, tokens)

	rate, err = Requests(requests, client, tokens, reporting)
	if err != nil {
		return []*http.Response{}, 0, err
	}

	resp := Report(len(requests), reporting)

	return resp, rate, nil
}
