// Package ratelimit implements API testing tool
// It enables load testing with rate-limiting and burst functionalities
package ratelimit

import (
	"net/http"
	"time"
)

// HTTPClient is an HTTP client that rate limits requests.
// If the rate is not specified, it defaults to a plain HTTP client.
type HTTPClient struct {
	http.Client
	rate   float64
	ticker *time.Ticker
}

// GetWithRateLimit issues a request.
func (c *HTTPClient) GetWithRateLimit(url string) (*http.Response, error) {
	if c.ticker != nil {
		<-c.ticker.C
	}
	return c.Get(url)
}

// DoWithRateLimit sends an http request.
func (c *HTTPClient) DoWithRateLimit(req *http.Request) (*http.Response, error) {
	if c.ticker != nil {
		<-c.ticker.C
	}

	return c.Do(req)
}

// NewHTTPClient returns a rate limited http client.
func NewHTTPClient(rate float64) HTTPClient {
	c := HTTPClient{
		rate: rate,
	}

	if rate != 0.0 {
		tickInterval := time.Duration(1e9/c.rate) * time.Nanosecond
		c.ticker = time.NewTicker(tickInterval)
	}

	return c
}
