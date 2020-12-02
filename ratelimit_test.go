package ratelimit_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/tgirier/ratelimit"
)

func TestGet(t *testing.T) {
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

		c := ratelimit.HTTPClient{}
		c.Client = tc.client
		c.Rate = tc.expectedRate

		start := time.Now()

		wg.Add(len(tc.requests))

		for _, req := range tc.requests {
			go func(req string) {
				c.Get(req)
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

func TestDo(t *testing.T) {
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

		c := ratelimit.HTTPClient{}
		c.Client = ts.Client()
		c.Rate = tc.expectedRate

		start := time.Now()

		wg.Add(len(requests))

		for _, req := range requests {
			go func(req *http.Request) {
				c.Do(req)
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
