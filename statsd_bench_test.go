package main

import (
	"testing"
)

func BenchmarkIncrement(b *testing.B) {
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

	tags := []string{"dest_host:example.com", "dest_port:443"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Increment("connectivity.check", tags)
	}
}
