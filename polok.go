// Package polok implements API testing tool
// It enables load testing with rate-limiting and burst functionalities
package polok

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Pipeline represents a data pipeline that executes a given function on a given dataset.
// The pipeline execution can be rate limited.
type Pipeline struct {
	Rate         float64       // Rate at which the pipeline should be limited
	WorkerNumber int           // Number of workers to spin up to execute the given functionality
	Client       *http.Client  // Client to use for http calls
	done         chan struct{} // Done channel triggering pipleine cancellation
}

// Run spins up a pipeline and launches the execution
// Closing the done channel cancel the whole pipeline
func (p *Pipeline) Run(inputStream <-chan *http.Request) (responseStream <-chan *http.Response) {
	resultStream := make(chan *http.Response)
	p.done = make(chan struct{})

	go func() {
		defer close(resultStream)
		bucket := make(chan struct{})

		var workerResultSteams []<-chan *http.Response

		for i := 0; i < p.WorkerNumber; i++ {
			result := p.do(inputStream, bucket)
			workerResultSteams = append(workerResultSteams, result)
		}

		p.limit(bucket)

		mergedStream := p.responseStreamsMerge(workerResultSteams...)

		for {
			select {
			case <-p.done:
				return
			case res := <-mergedStream:
				resultStream <- res
			}
		}

	}()

	return resultStream
}

// Stop cancels a running pipeline.
// It closes the internal done channel of the pipeline.
// It cancels all goroutines belonging to this pipeline.
func (p *Pipeline) Stop() {
	close(p.done)
}

// Limit is responsible for enforcing a global rate limiting expressed in requests per seconds
// Limit consumes a given number of tokens from the bucket channel at a given rate.
// A token represents a single request.
func (p *Pipeline) limit(bucket <-chan struct{}) {
	go func() {
		tickInterval := time.Duration(1e9/p.Rate) * time.Nanosecond
		tick := time.Tick(tickInterval)

		for {
			<-tick
			<-bucket
		}
	}()
}

// Do makes HTTP requests.
// It listens to an incoming stream of requests and return a channel of corresponding responses.
func (p *Pipeline) do(inputStream <-chan *http.Request, bucket chan<- struct{}) (responseStream <-chan *http.Response) {
	resultStream := make(chan *http.Response)

	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}

	go func() {
		defer close(resultStream)
		for {
			select {
			case <-p.done:
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

// responseStreamsMerge merges multiple response streams into one single stream
func (p *Pipeline) responseStreamsMerge(inputStreams ...<-chan *http.Response) (responseStream <-chan *http.Response) {
	var wg sync.WaitGroup
	multiplexedStream := make(chan *http.Response)

	multiplex := func(stream <-chan *http.Response) {
		defer wg.Done()
		for {
			select {
			case <-p.done:
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
