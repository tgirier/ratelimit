package polok_test

import (
	"testing"

	"github.com/tgirier/polok"
)

func TestRequest(t *testing.T) {
	method := "GET"
	url := "http://www.google.com"
	want := 200

	w := polok.Worker{}

	got, err := w.Request(method, url)
	if err != nil {
		t.Fatalf("worker - %v", err)
	}

	if got != want {
		t.Fatalf("worker - got %d, want %d", got, want)
	}
}

func TestLimiter(t *testing.T) {
	expectedRate := float64(100)

	total := 20

	bucket := make(chan struct{}, 20)

	for i := 0; i < 20; i++ {
		bucket <- struct{}{}
	}

	l := polok.Limiter{
		Rate: expectedRate,
	}

	got, rate := l.Consume(total, bucket)

	if got != total {
		t.Fatalf("limiter - got %d requests, want %d", got, total)
	}
	if rate > expectedRate {
		t.Fatalf("limiter - rate %v, expected %v", rate, expectedRate)
	}
	if len(bucket) > 0 {
		t.Fatalf("limiter - remaining tokens in bucket %v", len(bucket))
	}
}
