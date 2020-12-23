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

func TestHttpMethodsWithRateLimit(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World!")
	}))
	defer ts.Close()

	client := ts.Client()
	noLimitThreshold := 100.0
	errorMargin := 0.05

	testCases := []struct {
		name   string
		method string
		url    string
		number int
		rate   float64
	}{
		{name: "DO 2 reqs 1 QPS", method: "DO", url: ts.URL, number: 2, rate: 1.0},
		{name: "GET 2 reqs 1 QPS", method: "GET", url: ts.URL, number: 2, rate: 1.0},
		{name: "HEAD 2 reqs 1 QPS", method: "HEAD", url: ts.URL, number: 2, rate: 1.0},
		{name: "POST 2 reqs 1 QPS", method: "POST", url: ts.URL, number: 2, rate: 1.0},
		{name: "POSTFORM 2 reqs 1 QPS", method: "POSTFORM", url: ts.URL, number: 2, rate: 1.0},
		{name: "DO no rate limit", method: "DO", url: ts.URL, number: 2, rate: 0.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := ratelimit.NewHTTPClient(tc.rate)
			c.Transport = client.Transport

			start := time.Now()

			for i := 0; i < tc.number; i++ {
				switch tc.method {
				case "DO":
					req, err := http.NewRequest("GET", tc.url, nil)
					if err != nil {
						t.Errorf("%s - %v", tc.name, err)
					}
					c.DoWithRateLimit(req)
				case "GET":
					c.GetWithRateLimit(tc.url)
				case "HEAD":
					c.HeadWithRateLimit(tc.url)
				case "POST":
					c.PostWithRateLimit(tc.url, "", nil)
				case "POSTFORM":
					c.PostFormWithRateLimit(tc.url, nil)
				default:
					t.Errorf("%s - invalid method %v", tc.name, tc.method)
				}
			}

			stop := time.Now()
			duration := stop.Sub(start).Seconds()
			effectiveRate := float64(tc.number) / duration

			if tc.rate == 0 && effectiveRate < noLimitThreshold {
				t.Errorf("%s - effective rate %f, no limit threshold %.2f", tc.name, effectiveRate, noLimitThreshold)
			}

			if tc.rate != 0 && effectiveRate > tc.rate {
				t.Errorf("%s - effective rate too high %f, expected %.2f", tc.name, effectiveRate, tc.rate)
			}

			if tc.rate != 0 && effectiveRate < (tc.rate-errorMargin) {
				t.Errorf("%s - effective rate too low %f, expected %.2f, error margin %f", tc.name, effectiveRate, tc.rate, errorMargin)
			}

		})
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
