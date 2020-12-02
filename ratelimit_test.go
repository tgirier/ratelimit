package ratelimit_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/tgirier/ratelimit"
)

func TestGetWithRateLimit(t *testing.T) {
	t.Parallel()

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World !")
	}))

	testCases := []struct {
		name         string
		requests     []string
		expectedRate float64
		client       *http.Client
	}{
		{name: "2 requests - 1 QPS", requests: []string{ts.URL, ts.URL}, expectedRate: 1.0, client: ts.Client()},
	}

	for _, tc := range testCases {
		var wg sync.WaitGroup

		c := ratelimit.NewHTTPClient(tc.expectedRate)
		c.Transport = ts.Client().Transport

		start := time.Now()

		wg.Add(len(tc.requests))

		for _, req := range tc.requests {
			go func(req string) {
				c.GetWithRateLimit(req)
				wg.Done()
			}(req)
		}

		wg.Wait()

		stop := time.Now()
		duration := stop.Sub(start).Seconds()
		effectiveRate := float64(len(tc.requests)) / duration

		if effectiveRate > tc.expectedRate {
			t.Fatalf("effective rate %.2f, expected %.2f", effectiveRate, tc.expectedRate)
		}
	}

}

func TestDoWithRateLimit(t *testing.T) {
	t.Parallel()

	type Request struct {
		Method string
		URL    string
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World !")
	}))

	testCases := []struct {
		name         string
		requests     []Request
		expectedRate float64
		client       *http.Client
	}{
		{name: "2 requests - 1 QPS", requests: []Request{{Method: "GET", URL: ts.URL}, {Method: "GET", URL: ts.URL}}, expectedRate: 1.0, client: ts.Client()},
	}

	for _, tc := range testCases {
		var wg sync.WaitGroup
		var requests []*http.Request

		for _, req := range tc.requests {
			httpReq, err := http.NewRequest(req.Method, req.URL, nil)
			if err != nil {
				t.Fatalf("%v - %v", tc.name, err)
			}
			requests = append(requests, httpReq)
		}

		c := ratelimit.NewHTTPClient(tc.expectedRate)
		c.Transport = ts.Client().Transport

		start := time.Now()

		wg.Add(len(requests))

		for _, req := range requests {
			go func(req *http.Request) {
				c.DoWithRateLimit(req)
				wg.Done()
			}(req)
		}

		wg.Wait()

		stop := time.Now()
		duration := stop.Sub(start).Seconds()
		effectiveRate := float64(len(tc.requests)) / duration

		if effectiveRate > tc.expectedRate {
			t.Fatalf("effective rate %.2f, expected %.2f", effectiveRate, tc.expectedRate)
		}
	}

}

func TestPostWithRateLimit(t *testing.T) {
	t.Parallel()

	type Request struct {
		url         string
		contentType string
		body        io.Reader
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World !")
	}))

	testCases := []struct {
		name         string
		requests     []Request
		expectedRate float64
		client       *http.Client
	}{
		{name: "2 requests - 1 QPS", requests: []Request{{url: ts.URL, contentType: "", body: nil}, {url: ts.URL, contentType: "", body: nil}}, expectedRate: 1.0, client: ts.Client()},
	}

	for _, tc := range testCases {
		var wg sync.WaitGroup

		c := ratelimit.NewHTTPClient(tc.expectedRate)
		c.Transport = ts.Client().Transport

		start := time.Now()

		wg.Add(len(tc.requests))

		for _, req := range tc.requests {
			go func(req Request) {
				c.PostWithRateLimit(req.url, req.contentType, req.body)
				wg.Done()
			}(req)
		}

		wg.Wait()

		stop := time.Now()
		duration := stop.Sub(start).Seconds()
		effectiveRate := float64(len(tc.requests)) / duration

		if effectiveRate > tc.expectedRate {
			t.Fatalf("effective rate %.2f, expected %.2f", effectiveRate, tc.expectedRate)
		}
	}

}
