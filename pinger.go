package main

import (
	"fmt"
	"net"

	probing "github.com/prometheus-community/pro-bing"
)

func Ping(route *Route, dest *Destination, ip net.IP) bool {
	pinger, err := probing.NewPinger(ip.String())
	if err != nil {
		dest.Increment("connectivity.icmp.error", []string{})
		LogRouteError(route, fmt.Sprintf("Failed to setup ping to %s", ip.String()), err)
		return false
	}
	pinger.Count = 1
	err = pinger.Run()
	if err != nil {
		dest.Increment("connectivity.icmp.error", []string{})
		LogRouteError(route, fmt.Sprintf("Failed to ping %s", ip.String()), err)
		return false
	}

	// Emit metrics
	stats := pinger.Statistics()
	dest.Increment("connectivity.icmp.success", []string{})
	dest.Timer("connectivity.icmp", stats.AvgRtt, []string{})

	return true
}
