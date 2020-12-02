// Package ratelimit implements API testing tool
// It enables load testing with rate-limiting and burst functionalities
package ratelimit

import (
	"io"
	"net/http"
	"net/url"
	"time"
)

// HTTPClient is an HTTP client that rate limits requests.
// If the rate is not specified, it defaults to a plain HTTP client.
type HTTPClient struct {
	http.Client
	ticker *time.Ticker
}

// DoWithRateLimit issues a rate limited do request.
// All requests issued by this client using RateLimit methods share a common rate limiter.
// Those reuqests are waiting for an available tick from a ticker channel.
func (c *HTTPClient) DoWithRateLimit(req *http.Request) (resp *http.Response, err error) {
	if c.ticker != nil {
		<-c.ticker.C
	}
	return c.Do(req)
}

// GetWithRateLimit issues a rate lmited get request.
// All requests issued by this client using RateLimit methods share a common rate limiter.
// Those reuqests are waiting for an available tick from a ticker channel.
func (c *HTTPClient) GetWithRateLimit(url string) (resp *http.Response, err error) {
	if c.ticker != nil {
		<-c.ticker.C
	}
	return c.Get(url)
}

// HeadWithRateLimit issues a rate lmited head request.
// All requests issued by this client using RateLimit methods share a common rate limiter.
// Those reuqests are waiting for an available tick from a ticker channel.
func (c *HTTPClient) HeadWithRateLimit(url string) (resp *http.Response, err error) {
	if c.ticker != nil {
		<-c.ticker.C
	}
	return c.Head(url)
}

// PostWithRateLimit issues a rate limited post request.
// All requests issued by this client using RateLimit methods share a common rate limiter.
// Those reuqests are waiting for an available tick from a ticker channel.
func (c *HTTPClient) PostWithRateLimit(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	if c.ticker != nil {
		<-c.ticker.C
	}
	return c.Post(url, contentType, body)
}

// PostFormWithRateLimit issues a rate limited post form request.
// All requests issued by this client using RateLimit methods share a common rate limiter.
// Those reuqests are waiting for an available tick from a ticker channel.
func (c *HTTPClient) PostFormWithRateLimit(url string, data url.Values) (resp *http.Response, err error) {
	if c.ticker != nil {
		<-c.ticker.C
	}
	return c.PostForm(url, data)
}

// NewHTTPClient returns a rate limited http client.
func NewHTTPClient(rate float64) HTTPClient {
	c := HTTPClient{}

	if rate != 0.0 {
		tickInterval := time.Duration(1e9/rate) * time.Nanosecond
		c.ticker = time.NewTicker(tickInterval)
	}

	return c
}
