// Package polok implements API testing tool
// It enables load testing with rate-limiting and burst functionalities
package polok

import (
	"net/http"
	"time"
)

// RateLimitedHTTPClient is an HTTP client that rate limits requests.
// If the rate is not specified, it defaults to a plain HTTP client.
type RateLimitedHTTPClient struct {
	Transport         http.RoundTripper
	CheckRedirect     func(req *http.Request, via []*http.Request) error
	Jar               http.CookieJar
	Timeout           time.Duration
	Rate              float64
	ticker            *time.Ticker
	tickerInitialized bool
	clientInitialized bool
	client            http.Client
}

// Get issues a request.
func (c *RateLimitedHTTPClient) Get(url string) (*http.Response, error) {
	c.initializeTicker()
	c.initializeClient()

	if c.ticker != nil {
		<-c.ticker.C
	}

	return c.client.Get(url)
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

	c.client = http.Client{
		Transport:     c.Transport,
		CheckRedirect: c.CheckRedirect,
		Jar:           c.Jar,
		Timeout:       c.Timeout,
	}

	c.clientInitialized = true
}
