package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"time"

	"github.com/tgirier/ratelimit/proxy"
)

func main() {
	var wg sync.WaitGroup

	rate := 1.0 // Rate: 1.0 query per second
	n := 2      // Number of requests to be sent concurrently to the proxy

	// Create a backend server for which request will be proxied at a given rate
	backend := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "Request forwarded by the proxy reached the backend")
	}))
	defer backend.Close()
	fmt.Println("Backend server initialized")

	// Determine the URL of the backend
	urlToProxy, err := url.Parse(backend.URL)
	if err != nil {
		log.Fatal(err)
	}

	// Setup the proxy handler
	proxy := proxy.NewRateLimitedSingleRP(urlToProxy, rate)
	proxy.Server.Transport = backend.Client().Transport // Customizing the transport to ensure TLS trust to the backend

	// Create a frontend server using the proxy handler
	frontend := httptest.NewTLSServer(proxy)
	defer frontend.Close()
	fmt.Println("Frontend server acting as proxy initialized")

	// Start a timer to calculate the effective rate
	start := time.Now()
	fmt.Printf("Starting to proxy request to the backend at a rate of %.2f QPS\n", rate)

	wg.Add(n)

	// Send requests concurrently to the fronted server acting as a proxy
	for i := 0; i < n; i++ {
		go func() {
			resp, err := frontend.Client().Get(frontend.URL) // Using the frontend generated client to ensure TLS trust to the frontend
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()
			fmt.Println("Request sent")

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(b))

			wg.Done()
		}()
	}

	wg.Wait()

	// Calculate the effective rate
	stop := time.Now()
	duration := stop.Sub(start).Seconds()
	effectiveRate := float64(n) / duration

	fmt.Printf("Proxied %d requests at an effective rate of %.2f QPS\n", n, effectiveRate)
}
