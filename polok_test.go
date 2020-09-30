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
