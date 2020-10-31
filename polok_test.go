package polok_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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

	w := polok.Worker{
		Client: ts.Client(),
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("worker - %v", err)
	}

	resp, err := w.Request(req)
	if err != nil {
		t.Fatalf("worker - %v", err)
	}

	got := resp.StatusCode

	if got != want {
		t.Fatalf("worker - got %d, want %d", got, want)
	}
}

func TestMaxQPS(t *testing.T) {
	expectedRate := float64(100)

	total := 20

	bucket := make(chan struct{}, 20)

	for i := 0; i < 20; i++ {
		bucket <- struct{}{}
	}

	l := polok.MaxQPS{
		Rate: expectedRate,
	}

	got, rate := l.Consume(total, bucket)

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
