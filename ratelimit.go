// Package ratelimit implements rate limiting functionalities
// It defines types and rate limited associated methods
package ratelimit

import (
	"io"
	"net/http"
	"net/url"
	"time"
)

// HTTPClient is an HTTP client that rate limits requests.
// If the provided rate is zero, it defaults to a plain HTTP client.
type httpClient struct {
	http.Client
	ticker *time.Ticker
}

// DoWithRateLimit issues a rate limited do request.
// All requests issued by this client using RateLimit methods share a common rate limiter.
// Those reuqests are waiting for an available tick from a ticker channel.
func (c *httpClient) DoWithRateLimit(req *http.Request) (resp *http.Response, err error) {
	if c.ticker != nil {
		<-c.ticker.C
	}
	return c.Do(req)
}

// GetWithRateLimit issues a rate lmited get request.
// All requests issued by this client using RateLimit methods share a common rate limiter.
// Those reuqests are waiting for an available tick from a ticker channel.
func (c *httpClient) GetWithRateLimit(url string) (resp *http.Response, err error) {
	if c.ticker != nil {
		<-c.ticker.C
	}
	return c.Get(url)
}

// HeadWithRateLimit issues a rate lmited head request.
// All requests issued by this client using RateLimit methods share a common rate limiter.
// Those reuqests are waiting for an available tick from a ticker channel.
func (c *httpClient) HeadWithRateLimit(url string) (resp *http.Response, err error) {
	if c.ticker != nil {
		<-c.ticker.C
	}
	return c.Head(url)
}

// PostWithRateLimit issues a rate limited post request.
// All requests issued by this client using RateLimit methods share a common rate limiter.
// Those reuqests are waiting for an available tick from a ticker channel.
func (c *httpClient) PostWithRateLimit(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	if c.ticker != nil {
		<-c.ticker.C
	}
	return c.Post(url, contentType, body)
}

// PostFormWithRateLimit issues a rate limited post form request.
// All requests issued by this client using RateLimit methods share a common rate limiter.
// Those reuqests are waiting for an available tick from a ticker channel.
func (c *httpClient) PostFormWithRateLimit(url string, data url.Values) (resp *http.Response, err error) {
	if c.ticker != nil {
		<-c.ticker.C
	}
	return c.PostForm(url, data)
}

// NewHTTPClient returns a rate limited http client.
func NewHTTPClient(rate float64) *httpClient {
	c := httpClient{}

	if rate != 0.0 {
		tickInterval := time.Duration(1e9/rate) * time.Nanosecond
		c.ticker = time.NewTicker(tickInterval)
	}

	return &c
}

// Worker executes a given function at a given rate.
// If the provided rate is zero, it defaults to the provided function.
type worker struct {
	ticker *time.Ticker
	do     func()
}

// DoWithRateLimit executes the worker functionality at a given rate.
// All function exectued by this worker shares a common rate limiter.
// Those requests are waiting for an available tick from a ticker channel.
func (w *worker) DoWithRateLimit() {
	if w.ticker != nil {
		<-w.ticker.C
	}
	w.do()
}

// NewWorker returns a rate limited worker
func NewWorker(rate float64, f func()) *worker {
	w := worker{
		do: f,
	}

	if rate != 0.0 {
		tickInterval := time.Duration(1e9/rate) * time.Nanosecond
		w.ticker = time.NewTicker(tickInterval)
	}

	return &w
}
