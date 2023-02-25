package main

import (
	"log"
	"net"

	probing "github.com/prometheus-community/pro-bing"
)

func Ping(route *Route, dest *Destination, ip net.IP) bool {
	pinger, err := probing.NewPinger(ip.String())
	if err != nil {
		log.Printf("%s Failed to setup ping to %s (%v)", route, ip.String(), err)
		return false
	}
	pinger.Count = 1
	err = pinger.Run()
	if err != nil {
		log.Printf("%s Failed to ping %s (%v)", route, ip.String(), err)
		return false
	}

	stats := pinger.Statistics()
	dest.Timer("connectivity.icmp", stats.AvgRtt)

	return true
}
