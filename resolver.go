package main

import (
	"log"
	"net"
)

// Perform domain name resolution for a given destination, returning a list of
// IPs. If resolution is not successful, the list will be empty.
func Lookup(dest *Destination) []net.IP {
	var ips []net.IP

	results, err := net.LookupIP(dest.Host)
	if err != nil {
		log.Printf("%s Failed to resolve %s (%v)", GetLocalIPs(), dest.Host, err)
		return ips
	}

	for _, ip := range results {
		// Ignore IPv6 for now
		if ip.To4() != nil {
			ips = append(ips, ip)
		}
	}
	return ips
}
