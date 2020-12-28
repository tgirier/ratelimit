package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/tgirier/ratelimit/proxy"
)

func main() {

	// In this example, two mechanisms will be displayed.
	// The first one illustrates the standard setup of the proxy using urls passed to the constructor.
	// For this one plain http backends will be used.
	// The second one illustrates how you can customize the proxy handler by setting up manually the embedded router.
	// for this one, a tls backend will be used.

	rate := 1.0 // Rate at which the request will be proxied.

	// Create multiple http backends to display proxy standard setup.
	httpHostsNumber := 2
	urls, httpBackends, err := startMultipleServers(httpHostsNumber)
	if err != nil {
		closeMultipleSrvs(httpBackends)
		log.Fatal(err)
	}
	defer closeMultipleSrvs(httpBackends)
	fmt.Println("HTTP backends initialized")

	// Create the tls backend used to display proxy customization.
	tlsBackend := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello from tls backend")
	}))
	defer tlsBackend.Close()
	fmt.Println("HTTPS backend initialized")

	// Setup the proxy handler (standard)
	proxy := proxy.NewRateLimitedMultipleRP(rate, urls...)
	fmt.Println("Proxy handler initialized")

	// Customize the the proxy handler
	pattern := "/tls"
	tlsBackendURL, err := url.Parse(tlsBackend.URL)
	if err != nil {
		log.Fatal(err)
	}
	handler := httputil.NewSingleHostReverseProxy(tlsBackendURL)
	handler.Transport = tlsBackend.Client().Transport
	proxy.Router.Handle(pattern, handler)
	fmt.Println("Proxy handler customized")

	// Add the tls backend url to the list of url to request
	urls = append(urls, tlsBackendURL)

	// Create a frontend server usting the proxy handler
	frontend := httptest.NewTLSServer(proxy)
	defer frontend.Close()
	fmt.Println("Frontend server using the proxy handler initialized")

	// Start a timer to calculate the effective rate
	fmt.Printf("Starting to send requests to backends at a rate of %.2f QPS\n", rate)
	start := time.Now()

	// Send requests to the http backends
	for _, url := range urls {
		// Determine the URL to request to be proxied to the right proxy.
		// Hosts setup using standard proxy setup, the URL is as followed :
		// frontend_URL/backend_host
		// Ex: http://127.0.0.1:5001/www.google.com
		requestURL := frontend.URL + "/" + url.Host

		// If the request has to be proxied to the tls backend, use the customized path instead.
		if url.Scheme == "https" {
			requestURL = frontend.URL + "/tls"
		}

		resp, err := frontend.Client().Get(requestURL)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		fmt.Println("Request sent to: ", requestURL)

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	}

	// Calculate the effective rate
	stop := time.Now()
	duration := stop.Sub(start).Seconds()
	effectiveRate := float64(len(urls)) / duration
	fmt.Printf("Proxied %d requests at an effective rate of %.2f QPS\n", len(urls), effectiveRate)
}

// Function use to spin up multiple http servers.
func startMultipleServers(n int) ([]*url.URL, []*httptest.Server, error) {

	var srvs []*httptest.Server
	var urls []*url.URL

	for i := 0; i < n; i++ {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "Hello from backend ", r.URL.Path[1:])
		}))
		srvs = append(srvs, ts)

		url, err := url.Parse(ts.URL)
		if err != nil {
			return nil, srvs, err
		}
		urls = append(urls, url)
	}

	return urls, srvs, nil
}

// Function used to close multiple servers
func closeMultipleSrvs(srvs []*httptest.Server) {
	for _, srv := range srvs {
		srv.Close()
	}
}
