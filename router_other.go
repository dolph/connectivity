//go:build !linux

package main

import (
	"fmt"
	"net"
	"runtime"
)

func lookupRoute(ip net.IP, route *Route) error {
	return fmt.Errorf("routing table lookup not supported on %s", runtime.GOOS)
}
