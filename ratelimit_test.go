package ratelimit_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/tgirier/ratelimit"
)

func TestHttpDoWithRateLimit(t *testing.T) {
	t.Parallel()

	type Request struct {
		Method string
		URL    string
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World !")
	}))
	defer ts.Close()

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

func TestHttpGetWithRateLimit(t *testing.T) {
	t.Parallel()

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World !")
	}))
	defer ts.Close()

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

func TestHttpHeadWithRateLimit(t *testing.T) {
	t.Parallel()

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World !")
	}))
	defer ts.Close()

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
				c.HeadWithRateLimit(req)
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

func TestHttpPostWithRateLimit(t *testing.T) {
	t.Parallel()

	type Request struct {
		url         string
		contentType string
		body        io.Reader
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World !")
	}))
	defer ts.Close()

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

func TestHttpPostFormWithRateLimit(t *testing.T) {
	t.Parallel()

	type Request struct {
		url  string
		data url.Values
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World !")
	}))
	defer ts.Close()

	testCases := []struct {
		name         string
		requests     []Request
		expectedRate float64
		client       *http.Client
	}{
		{name: "2 requests - 1 QPS", requests: []Request{{url: ts.URL, data: nil}, {url: ts.URL, data: nil}}, expectedRate: 1.0, client: ts.Client()},
	}

	for _, tc := range testCases {
		var wg sync.WaitGroup

		c := ratelimit.NewHTTPClient(tc.expectedRate)
		c.Transport = ts.Client().Transport

		start := time.Now()

		wg.Add(len(tc.requests))

		for _, req := range tc.requests {
			go func(req Request) {
				c.PostFormWithRateLimit(req.url, req.data)
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

func TestWorkerDoWithRateLimit(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		expectedRate float64
		f            func()
		n            int
	}{
		{name: "2 requests - 1 QPS", expectedRate: 1.0, f: func() { fmt.Println("Hello") }, n: 2},
	}

	for _, tc := range testCases {
		var wg sync.WaitGroup

		w := ratelimit.NewWorker(tc.expectedRate, tc.f)

		start := time.Now()

		wg.Add(tc.n)

		for i := 0; i < tc.n; i++ {
			go func() {
				w.DoWithRateLimit()
				wg.Done()
			}()
		}

		wg.Wait()

		stop := time.Now()
		duration := stop.Sub(start).Seconds()
		effectiveRate := float64(tc.n) / duration

		if effectiveRate > tc.expectedRate {
			t.Fatalf("effective rate %.2f, expected %.2f", effectiveRate, tc.expectedRate)
		}
	}
}
