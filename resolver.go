package main

import (
	"log"
	"net"
)

func Lookup(dest *Destination) []string {
	var ips []string

	results, err := net.LookupIP(dest.Host)
	if err != nil {
		log.Printf("%s Failed to resolve %s (%v)", GetLocalIPs(), dest.Host, err)
		return ips
	}

	for _, ip := range results {
		s := ip.String()
		ips = append(ips, s)
	}
	return ips
}
