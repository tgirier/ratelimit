package polok_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/tgirier/polok"
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

		c := polok.RateLimitedHTTPClient{
			Transport: tc.client.Transport,
		}
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
