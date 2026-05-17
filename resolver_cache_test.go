package main

import (
	"net"
	"testing"
)

func TestLookupUsesDNSCache(t *testing.T) {
	clearDNSCache()
	t.Cleanup(clearDNSCache)

	dest, err := NewDestination(Url{Label: "localhost", Url: "http://127.0.0.1"})
	if err != nil {
		t.Fatal(err)
	}

	first, err := Lookup(dest)
	if err != nil {
		t.Fatalf("first Lookup: %v", err)
	}

	if _, ok := dnsCache.Load(dest.Host); !ok {
		t.Fatal("expected cache entry after first Lookup")
	}

	second, err := Lookup(dest)
	if err != nil {
		t.Fatalf("second Lookup: %v", err)
	}

	if len(first) != len(second) {
		t.Fatalf("len first=%d second=%d", len(first), len(second))
	}
	for i := range first {
		if first[i].String() != second[i].String() {
			t.Fatalf("ip[%d] first=%s second=%s", i, first[i], second[i])
		}
	}

	// Cached slice must not be shared with callers.
	first[0] = net.ParseIP("0.0.0.0")
	third, err := Lookup(dest)
	if err != nil {
		t.Fatalf("third Lookup: %v", err)
	}
	if third[0].String() == "0.0.0.0" {
		t.Error("Lookup returned mutable cached slice")
	}
}
