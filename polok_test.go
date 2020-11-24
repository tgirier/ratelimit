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
	t.Parallel()
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
	defer p.Stop()

	input := make(chan *http.Request, numRequests)

	for i := 0; i < numRequests; i++ {
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			t.Fatal(err)
		}
		input <- req
	}

	start := time.Now()

	results := p.Run(input)

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

func TestSlicetoStream(t *testing.T) {
	t.Parallel()
	var inputRequests []*http.Request
	numRequests := 3

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World !")
	}))

	method := "GET"
	url := ts.URL

	for i := 0; i < numRequests; i++ {
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			t.Fatal(err)
		}
		inputRequests = append(inputRequests, req)
	}

	resultStream := polok.SlicetoStream(inputRequests...)

	for i := 0; i < numRequests; i++ {
		<-resultStream
	}

	if len(resultStream) != 0 {
		t.Fatalf("%v remaining messages in result stream", len(resultStream))
	}
}

func TestStreamtoSlice(t *testing.T) {
	t.Parallel()
	numResponse := 3
	stream := make(chan *http.Response, numResponse)

	for i := 0; i < numResponse; i++ {
		res, err := newResponse()
		if err != nil {
			t.Fatal(err)
		}
		stream <- res
	}

	out := polok.StreamtoSlice(stream, numResponse)

	if len(out) != numResponse {
		t.Fatalf("got slice of length %v, want %v", len(out), numResponse)
	}
}

func newResponse() (*http.Response, error) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello World !")
	}))
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ts.Client().Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
