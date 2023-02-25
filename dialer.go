package main

import (
	"fmt"
	"log"
	"net"
)

// Try to open a connection to the destination, and then immediately disconnect
// if succcessful. This ensures we have a network path to the destination, and
// validates each individual record in the DNS response.
func Dial(dest *Destination, ip net.IP) bool {
	metricTags := []string{fmt.Sprintf("dest_ip:%s", ip.String())}
	hostPort := fmt.Sprintf("%s:%d", ip.String(), dest.Port)

	// Check destination IP for routability
	route, err := GetRoute(ip)
	if err != nil {
		log.Printf("%s Failed to route to %s: %v", dest, ip.String(), err)
		return false
	}

	// Test destination IP by dialing route
	dest.Increment("connectivity.dial", metricTags)
	conn, err := net.Dial(dest.Protocol, hostPort)
	if err != nil {
		dest.Increment("connectivity.dial.error", metricTags)
		log.Printf("%s%s Failed to %v", route, dest, err)
		return false
	}
	defer conn.Close()
	return true
}
