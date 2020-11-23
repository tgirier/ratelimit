package polok_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tgirier/polok"
)

func TestRequest(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello World")
	}))
	defer ts.Close()

	method := "GET"
	url := ts.URL
	want := http.StatusOK
	var wg sync.WaitGroup

	bucket := make(chan struct{}, 1)
	reporting := make(chan *http.Response, 1)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("worker - %v", err)
	}

	wg.Add(1)

	err = polok.Request(req, ts.Client(), bucket, reporting, &wg)
	if err != nil {
		t.Fatalf("worker - %v", err)
	}

	resp := <-reporting

	got := resp.StatusCode

	if got != want {
		t.Fatalf("worker - got %d, want %d", got, want)
	}
	if len(bucket) == 0 {
		t.Fatalf("worker - remaining tokens in bucket %v", len(bucket))
	}
}

func TestNewRequest(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello World !")
	}))
	defer ts.Close()

	method := "GET"
	url := ts.URL
	want := http.StatusOK
	done := make(chan struct{})
	defer close(done)

	bucket := make(chan struct{}, 1)
	input := make(chan *http.Request, 1)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("request - %v", err)
	}

	input <- req

	result := polok.NewRequester(done, input, bucket, ts.Client())

	res := <-result
	got := res.StatusCode

	if got != want {
		t.Fatalf("new request - got %v, want %v", got, want)
	}
	if len(result) != 0 {
		t.Fatalf("new request - %v responses left in channel", len(result))
	}
}

func TestResponseStreamMerge(t *testing.T) {
	var streams []<-chan *http.Response

	streamsNumber := 2

	done := make(chan struct{})
	defer close(done)

	for i := 0; i < streamsNumber; i++ {
		stream := make(chan *http.Response, 1)
		res, err := newResponse()
		if err != nil {
			t.Fatal(err)
		}
		stream <- res
		streams = append(streams, stream)
	}

	result := polok.ResponseStreamsMerge(done, streams...)

	for i := 0; i < streamsNumber; i++ {
		<-result
	}

	if len(result) != 0 {
		t.Fatalf("got %v messages left in merged channel", len(result))
	}
}

func TestPipelineDo(t *testing.T) {
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

	results := p.Do(done, input)

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

func TestReport(t *testing.T) {
	n := 2
	reporting := make(chan *http.Response, n)
	want := "Hello World !"

	for i := 0; i < n; i++ {
		resp, err := newResponse()
		if err != nil {
			t.Fatal(err)
		}
		reporting <- resp
	}

	responses := polok.Report(n, reporting)

	for _, resp := range responses {

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		got := strings.TrimSuffix(string(body), "\n")

		if got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
	}

	if len(responses) != n {
		t.Fatalf("expected %d responses, got %d", n, len(responses))
	}

}

func TestRequestWithLimit(t *testing.T) {
	initialRate := 100.0
	requestNumber := 20
	var requests []*http.Request

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello World !")
	}))
	defer ts.Close()

	for i := 0; i < requestNumber; i++ {
		req, err := http.NewRequest("GET", ts.URL, nil)
		if err != nil {
			t.Fatalf("request - %v", err)
		}
		requests = append(requests, req)
	}

	resp, rate, err := polok.RequestWithLimit(requests, initialRate, ts.Client())
	if err != nil {
		t.Fatalf("request with limit - %v", err)
	}

	got := len(resp)
	expectedRate := initialRate
	want := requestNumber

	if got != want {
		t.Fatalf("request - got %v requests, want %v", got, want)
	}
	if rate > expectedRate {
		t.Fatalf("request - rate %v, expected %v", rate, expectedRate)
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
