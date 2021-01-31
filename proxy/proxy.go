// Package proxy implement a rate limited HTTP proxy
package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// rateLimitedSingleRP is an http proxy that rate limits outgoing requests for a single host.
// If the provided rate is zero, it defaults to a plain http reverse proxy.
type rateLimitedSingleRP struct {
	Server httputil.ReverseProxy
	ticker *time.Ticker
}

// ServeHTTP is an http handler.
// It listens to incoming requests, waits for an available ticker and sends the request back to the initial caller.
func (p *rateLimitedSingleRP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p.ticker != nil {
		<-p.ticker.C
	}
	p.Server.ServeHTTP(w, r)
}

// NewRateLimitedSingleRP returns a rate limited http proxy for the given URL.
func NewRateLimitedSingleRP(rate float64, target *url.URL) *rateLimitedSingleRP {
	rp := httputil.NewSingleHostReverseProxy(target)

	p := &rateLimitedSingleRP{
		Server: *rp,
	}

	if rate != 0.0 {
		tickInterval := time.Duration(1e9/rate) * time.Nanosecond
		p.ticker = time.NewTicker(tickInterval)
	}

	return p
}

// rateLimitedMultipleRP is an http reverse proxy that rate limits outgoing requests for multiple hosts.
// The rate is globally enforced at the proxy level.
// If the provided rate is zero, it defaults to a plain http reverse proxy.
type rateLimitedMultipleRP struct {
	Router *http.ServeMux
	ticker *time.Ticker
}

// ServeHTTP is an http handler.
// It listens to incoming resquests and passes it to the embedded router at a given rate.
func (mp *rateLimitedMultipleRP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if mp.ticker != nil {
		<-mp.ticker.C
	}
	mp.Router.ServeHTTP(w, r)
}

// NewRateLimitedMultipleRP returns a multiple host rate limited reverse proxy.
func NewRateLimitedMultipleRP(rate float64, targets ...*url.URL) *rateLimitedMultipleRP {
	mp := &rateLimitedMultipleRP{}

	if rate != 0.0 {
		tickerInterval := time.Duration(1e9/rate) * time.Nanosecond
		mp.ticker = time.NewTicker(tickerInterval)
	}

	mp.Router = http.NewServeMux()

	for _, url := range targets {
		pattern := fmt.Sprint("/", url.Host)
		handler := httputil.NewSingleHostReverseProxy(url)
		mp.Router.Handle(pattern, handler)
	}

	return mp
}
