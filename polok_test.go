package polok_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tgirier/polok"
)

func TestPipelineRun(t *testing.T) {
	expectedRate := 1.0
	numWorkers := 2
	numRequests := 3

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello World!")
	}))

	method := "GET"
	url := ts.URL
	want := http.StatusOK

	p := polok.Pipeline{
		Rate:         expectedRate,
		WorkerNumber: numWorkers,
		Client:       ts.Client(),
	}

	input := make(chan *http.Request, numRequests)
	done := make(chan struct{})
	defer close(done)

	for i := 0; i < numRequests; i++ {
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			t.Fatal(err)
		}
		input <- req
	}

	start := time.Now()

	results := p.Run(done, input)

	for i := 0; i < numRequests; i++ {
		res := <-results
		got := res.StatusCode
		if got != want {
			t.Fatalf("got %v, want %v", got, want)
		}
	}

	stop := time.Now()
	duration := stop.Sub(start).Seconds()

	rate := float64(numRequests) / duration

	if len(results) != 0 {
		t.Fatalf("remaining %d messages in results channel", len(results))
	}
	if rate > expectedRate {
		t.Fatalf("rate %v, expected %v", rate, expectedRate)
	}
}
