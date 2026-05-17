package main

import (
	"testing"
)

func BenchmarkDestinationIncrement(b *testing.B) {
	dest, err := NewDestination(Url{Label: "x", Url: "https://example.com"})
	if err != nil {
		b.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-queue:
			case <-done:
				return
			}
		}
	}()
	defer close(done)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dest.Increment("connectivity.check", nil)
	}
}
