package main

import (
	"fmt"
	"log"
	"net"
)

// Try to open a connection to the destination, and then immediately disconnect
// if succcessful. This ensures we have a network path to the destination, and
// can be used to ensure that individual records in a DNS response are available.
func Dial(dest *Destination, ip net.IP) bool {
	hostPort := fmt.Sprintf("%s:%d", ip.String(), dest.Port)

	// Check destination IP for routability
	route, err := GetRoute(ip)
	if err != nil {
		log.Printf("Failed to route to %s (%v)", ip.String(), err)
		return false
	}

	// Test destination IP by dialing route
	conn, err := net.Dial(dest.Protocol, hostPort)
	if err != nil {
		log.Printf("%s Failed to %v", route, err)
		return false
	}
	defer conn.Close()
	return true
}
