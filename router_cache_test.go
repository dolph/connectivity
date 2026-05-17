package main

import (
	"net"
	"testing"
)

func TestLookupRouteReusesCachedRouter(t *testing.T) {
	clearRouteRouterForTest()
	t.Cleanup(clearRouteRouterForTest)

	ip := net.ParseIP("127.0.0.1")
	if _, err := GetRoute(ip); err != nil {
		t.Skipf("routing unavailable on this platform: %v", err)
	}

	routeRouterMu.RLock()
	first := routeRouter
	routeRouterMu.RUnlock()
	if first == nil {
		t.Fatal("expected cached router after first lookup")
	}

	if _, err := GetRoute(ip); err != nil {
		t.Fatalf("second GetRoute: %v", err)
	}

	routeRouterMu.RLock()
	second := routeRouter
	routeRouterMu.RUnlock()
	if first != second {
		t.Error("expected same routing.Router instance on second lookup")
	}
}
