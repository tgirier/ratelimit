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
	rate   float64
	client *http.Client
	ticker *time.Ticker
}

// Get issues a request.
func (c *HTTPClient) Get(url string) (*http.Response, error) {
	if c.client != nil {
		<-c.ticker.C
	}
	return c.client.Get(url)
}

// Do sends an http request.
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	if c.client != nil {
		<-c.ticker.C
	}

	return c.client.Do(req)
}

// NewHTTPClient returns a rate limited http client.
// The client can be configured by providing a pointer to an http client.
// If the provider pointer is nil Http default client will be used.
func NewHTTPClient(client *http.Client, rate float64) HTTPClient {
	c := HTTPClient{
		client: http.DefaultClient,
		rate:   rate,
	}

	if client != nil {
		c.client = client
	}

	if rate != 0.0 {
		tickInterval := time.Duration(1e9/c.rate) * time.Nanosecond
		c.ticker = time.NewTicker(tickInterval)
	}

	return c
}
