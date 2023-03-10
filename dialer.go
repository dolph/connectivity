package main

import (
	"fmt"
	"net"
)

// Try to open a connection to the destination, and then immediately disconnect
// if succcessful. This ensures we have a network path to the destination, and
// validates each individual record in the DNS response.
func Dial(route *Route, dest *Destination, ip net.IP) bool {
	metricTags := []string{fmt.Sprintf("dest_ip:%s", ip.String())}
	hostPort := fmt.Sprintf("%s:%d", ip.String(), dest.Port)

	// Test destination IP by dialing route
	dest.Increment("connectivity.dial", metricTags)
	conn, err := net.Dial(dest.Protocol, hostPort)
	if err != nil {
		dest.Increment("connectivity.dial.error", metricTags)
		LogRouteDestinationError(route, dest, "Failed", err)
		return false
	}
	defer conn.Close()
	dest.Increment("connectivity.dial.success", metricTags)
	return true
}
