package main

import "testing"

func BenchmarkGetLocalIPs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetLocalIPs()
	}
}
