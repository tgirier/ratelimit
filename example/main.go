package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/tgirier/polok"
)

func main() {
	var wg sync.WaitGroup
	var n int
	var rate float64
	reqs := []string{
		"https://google.com",
		"https://bitfieldconsulting.com",
	}
	m := polok.MaxQPS{
		Rate: 0.2,
	}
	w := polok.Worker{}
	tokens := make(chan struct{})
	go func() {
		n, rate = m.Consume(len(reqs), tokens)
	}()
	wg.Add(len(reqs))
	for _, URL := range reqs {
		go func(u string) {
			tokens <- struct{}{}
			fmt.Println("requesting ", u)
			_, _ = w.Request(http.MethodGet, u)
			wg.Done()
		}(URL)
	}
	wg.Wait()
	fmt.Printf("total requests: %d, average rate: %.2f\n", n, rate)
}