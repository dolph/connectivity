package main

import (
	"net"
)

// Perform domain name resolution for a given destination, returning a list of
// IPs. If resolution is not successful, the list will be empty.
func Lookup(dest *Destination) ([]net.IP, error) {
	results, err := net.LookupIP(dest.Host)
	if err != nil {
		return nil, err
	}

	var ips []net.IP
	for _, ip := range results {
		// Ignore IPv6 for now
		if ip.To4() != nil {
			ips = append(ips, ip)
		}
	}
	return ips, nil
}
