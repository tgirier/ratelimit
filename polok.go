// Package polok implements API testing tool
// It enables load testing with rate-limiting and burst functionalities
package polok

import (
	"net/http"
	"time"
)

type httpClient interface {
	Get(string) (*http.Response, error)
	Do(*http.Request) (*http.Response, error)
}

// RateLimitedHTTPClient is an HTTP client that rate limits requests.
// If the rate is not specified, it defaults to a plain HTTP client.
type RateLimitedHTTPClient struct {
	Client            httpClient
	Rate              float64
	ticker            *time.Ticker
	tickerInitialized bool
	clientInitialized bool
}

// Get issues a request.
func (c *RateLimitedHTTPClient) Get(url string) (*http.Response, error) {
	c.initializeTicker()
	c.initializeClient()

	c.rateLimit()

	return c.Client.Get(url)
}

// Do sends an http request.
func (c *RateLimitedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.initializeTicker()
	c.initializeClient()

	c.rateLimit()

	return c.Client.Do(req)
}

// initializeTicker sets up a client time ticker base on the provided rate.
func (c *RateLimitedHTTPClient) initializeTicker() {
	if c.tickerInitialized {
		return
	}
	if c.ticker == nil && c.Rate != 0.0 {
		tickInterval := time.Duration(1e9/c.Rate) * time.Nanosecond
		c.ticker = time.NewTicker(tickInterval)
	}
	c.tickerInitialized = true
}

// initializeClient configures the embedded http client based on specified propoerties.
func (c *RateLimitedHTTPClient) initializeClient() {
	if c.clientInitialized {
		return
	}

	if c.Client == nil {
		c.Client = http.DefaultClient
	}

	c.clientInitialized = true
}

// rateLimit wait for a tick if the client is rate limited.
func (c *RateLimitedHTTPClient) rateLimit() {
	if c.ticker != nil {
		<-c.ticker.C
	}
}
