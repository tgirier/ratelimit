// Package polok implements API testing tool
// It enables load testing with rate-limiting and burst functionalities
package polok

import (
	"net/http"
)

// Worker is responsible for requesting a given URL with a given method
type Worker struct{}

// Request makes a request to a given URL with a given method
func (w *Worker) Request(method string, url string) (int, error) {
	c := http.Client{}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return 0, err
	}

	sc := resp.StatusCode

	return sc, nil
}
