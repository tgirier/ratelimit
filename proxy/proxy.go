// Package proxy implement a rate limited HTTP proxy
package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// RateLimitedRP is an http proxy that rate limits outgoing requests.
// If the provided rate is zero, it defaults to a plain http reverse proxy.
type RateLimitedRP struct {
	Server httputil.ReverseProxy
	ticker *time.Ticker
}

// ServeHTTP is an http handler.
// It listens to incoming requests, waits for an available ticker and sends the request back to the initial caller.
func (p *RateLimitedRP) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if p.ticker != nil {
		<-p.ticker.C
	}
	p.Server.ServeHTTP(rw, req)
}

// NewRateLimitedReverseProxy returns a rate limited http proxy for the given URL.
func NewRateLimitedReverseProxy(target *url.URL, rate float64) *RateLimitedRP {
	rp := httputil.NewSingleHostReverseProxy(target)

	p := &RateLimitedRP{
		Server: *rp,
	}

	if rate != 0.0 {
		tickInterval := time.Duration(1e9/rate) * time.Nanosecond
		p.ticker = time.NewTicker(tickInterval)
	}

	return p
}
