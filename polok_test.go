package polok_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

	bucket := make(chan struct{}, 1)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("worker - %v", err)
	}

	resp, err := polok.Request(req, bucket, ts.Client())
	if err != nil {
		t.Fatalf("worker - %v", err)
	}

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

	got := polok.Limit(total, expectedRate, bucket)

	stop := time.Now()
	duration := stop.Sub(start).Seconds()

	rate := float64(got) / duration

	if got != total {
		t.Fatalf("limiter - got %d requests, want %d", got, total)
	}
	if rate > expectedRate {
		t.Fatalf("limiter - rate %v, expected %v", rate, expectedRate)
	}
	if len(bucket) > 0 {
		t.Fatalf("limiter - remaining tokens in bucket %v", len(bucket))
	}
}

func TestRequestWithLimit(t *testing.T) {
	initialRate := 100.0
	requestNumber := 20

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello World !")
	}))
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatalf("request - %v", err)
	}

	got, rate, err := polok.RequestWithLimit(req, requestNumber, initialRate, ts.Client())

	expectedRate := initialRate
	want := requestNumber

	if got != want {
		t.Fatalf("request - got %v requests, want %v", got, want)
	}
	if rate > expectedRate {
		t.Fatalf("request - rate %v, expected %v", rate, expectedRate)
	}
}
