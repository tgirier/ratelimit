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
func Limit(rate float64, bucket <-chan struct{}) {
	go func() {
		tickInterval := time.Duration(1e9/rate) * time.Nanosecond
		tick := time.Tick(tickInterval)

		for {
			<-tick
			<-bucket
		}
	}()
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

// ResponseStreamsMerge merges multiple response streams into one single stream
func ResponseStreamsMerge(done <-chan struct{}, inputStreams ...<-chan *http.Response) (responseStream <-chan *http.Response) {
	var wg sync.WaitGroup
	multiplexedStream := make(chan *http.Response)

	multiplex := func(stream <-chan *http.Response) {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			case res := <-stream:
				multiplexedStream <- res
			}
		}
	}

	wg.Add(len(inputStreams))
	for _, stream := range inputStreams {
		go multiplex(stream)
	}

	go func() {
		wg.Wait()
		close(multiplexedStream)
	}()

	return multiplexedStream
}

// Pipeline represents a data pipeline that executes a given function on a given dataset.
// The pipeline execution can be rate limited.
type Pipeline struct {
	Rate         float64      // Rate at which the pipeline should be limited
	WorkerNumber int          // Number of workers to spin up to execute the given functionality
	Client       *http.Client // Client to use for http calls
}

// Do spins up a pipeline and launches the execution
// Closing the done channel cancel the whole pipeline
func (p *Pipeline) Do(done <-chan struct{}, inputStream <-chan *http.Request) (responseStream <-chan *http.Response) {
	resultStream := make(chan *http.Response)

	go func() {
		defer close(resultStream)
		bucket := make(chan struct{})

		var workerResultSteams []<-chan *http.Response

		for i := 0; i < p.WorkerNumber; i++ {
			result := NewRequester(done, inputStream, bucket, p.Client)
			workerResultSteams = append(workerResultSteams, result)
		}

		Limit(p.Rate, bucket)

		mergedStream := ResponseStreamsMerge(done, workerResultSteams...)

		for {
			select {
			case <-done:
				return
			case res := <-mergedStream:
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

	go Limit(rate, tokens)

	rate, err = Requests(requests, client, tokens, reporting)
	if err != nil {
		return []*http.Response{}, 0, err
	}

	resp := Report(len(requests), reporting)

	return resp, rate, nil
}
