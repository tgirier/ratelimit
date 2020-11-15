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

func TestLimit(t *testing.T) {
	expectedRate := float64(100)

	total := 20

	bucket := make(chan struct{}, 20)

	for i := 0; i < 20; i++ {
		bucket <- struct{}{}
	}

	start := time.Now()

	polok.Limit(total, expectedRate, bucket)

	stop := time.Now()
	duration := stop.Sub(start).Seconds()

	rate := float64(total) / duration

	if rate > expectedRate {
		t.Fatalf("limiter - rate %v, expected %v", rate, expectedRate)
	}
	if len(bucket) != 0 {
		t.Fatalf("limiter - remaining tokens in bucket %v", len(bucket))
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
