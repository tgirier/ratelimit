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

	testCases := []struct {
		name   string
		number int
		rate   float64
		want   string
	}{
		{name: "2 requests, 1 QPS", number: 2, rate: 1.0, want: "Request forwarded by proxy"},
	}

	for _, tc := range testCases {
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, tc.want)
		}))
		defer ts.Close()

		rpURL, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatal(err)
		}

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

			if s != tc.want {
				t.Fatalf("%s - got: %s, expected: %s", tc.name, s, tc.want)
			}
		}

		stop := time.Now()
		duration := stop.Sub(start).Seconds()
		effectiveRate := float64(tc.number) / duration

		if effectiveRate > tc.rate {
			t.Fatalf("effective rate %f, expected %.2f", effectiveRate, tc.rate)
		}
	}
}
