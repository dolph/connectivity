package main

import (
	"net"
	"time"
)

// Perform domain name resolution for a given destination, returning a list of
// IPs. If resolution is not successful, the list will be empty.
func Lookup(dest *Destination) ([]net.IP, error) {
	t1 := time.Now()
	ips, err := lookupHostIPv4(dest.Host)
	t2 := time.Now()
	dest.Timer("connectivity.lookup", t2.Sub(t1), []string{})

	if err != nil {
		dest.Increment("connectivity.lookup.error", []string{})
		return nil, err
	}

	dest.Increment("connectivity.lookup.success", []string{})
	return ips, nil
}
