package polok_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tgirier/polok"
)

func TestRequest(t *testing.T) {
	s, err := startTestTLSServer()
	if err != nil {
		t.Fatalf("worker - test tls server %v", err)
	}
	defer s.Close()

	method := "GET"
	url := s.URL
	want := http.StatusOK

	w := polok.Worker{}

	resp, err := w.Request(method, url)

	got := resp.StatusCode

	if err != nil {
		t.Fatalf("worker - %v", err)
	}

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

func startTestTLSServer() (*httptest.Server, error) {
	s := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello World")
	}))

	cert, err := tls.LoadX509KeyPair("testCerts/server.crt", "testCerts/server.key")
	if err != nil {
		return nil, err
	}

	s.TLS = &tls.Config{Certificates: []tls.Certificate{cert}}
	s.StartTLS()

	return s, nil
}
