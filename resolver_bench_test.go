package main

import (
	"testing"
)

func BenchmarkLookup(b *testing.B) {
	dest, err := NewDestination(Url{Label: "localhost", Url: "http://127.0.0.1"})
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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Lookup(dest); err != nil {
			b.Fatal(err)
		}
	}
}
