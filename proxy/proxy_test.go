package proxy_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/tgirier/ratelimit/proxy"
)

func TestServeHTTPWithRateLimit(t *testing.T) {
	t.Parallel()

	want := "Request forwarded by proxy"
	noLimitThreshold := 100.0
	errorMargin := 0.05

	testCases := []struct {
		name   string
		number int
		rate   float64
		want   string
	}{
		{name: "2 requests, 1 QPS", number: 2, rate: 1.0},
		{name: "2 requests, no rate limiting", number: 2, rate: 0.0},
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, want)
	}))
	defer ts.Close()

	rpURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		rp := proxy.NewRateLimitedReverseProxy(rpURL, tc.rate)
		rp.Server.Transport = ts.Client().Transport

		p := httptest.NewTLSServer(rp)
		defer p.Close()

		start := time.Now()

		for i := 0; i < tc.number; i++ {
			resp, err := p.Client().Get(p.URL)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			s := string(b)

			if s != want {
				t.Fatalf("%s - got: %s, expected: %s", tc.name, s, want)
			}
		}

		stop := time.Now()
		duration := stop.Sub(start).Seconds()
		effectiveRate := float64(tc.number) / duration

		if tc.rate == 0.0 && effectiveRate < noLimitThreshold {
			t.Fatalf("%s - effective rate %f, no rate limiting threshold %.2f", tc.name, effectiveRate, noLimitThreshold)
		}

		if tc.rate != 0.0 && effectiveRate > tc.rate {
			t.Fatalf("%s - effective rate too high %f, expected %.2f", tc.name, effectiveRate, tc.rate)
		}

		if tc.rate != 0.0 && effectiveRate < (tc.rate-errorMargin) {
			t.Fatalf("%s - effective rate too low %f, expected %.2f, error margin %f", tc.name, effectiveRate, tc.rate, errorMargin)
		}
	}
}
