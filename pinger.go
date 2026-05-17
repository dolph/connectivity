package main

import (
	"fmt"
	"net"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

const (
	icmpProbeCount     = 3
	icmpProbeInterval  = 200 * time.Millisecond
)

// icmpProbeSucceeded reports whether enough probe replies arrived.
func icmpProbeSucceeded(packetsRecv int, probeCount int) bool {
	return packetsRecv >= (probeCount+1)/2
}

func Ping(route *Route, dest *Destination, ip net.IP) bool {
	pinger, err := probing.NewPinger(ip.String())
	if err != nil {
		dest.Increment("connectivity.icmp.error", []string{})
		LogRouteError(route, fmt.Sprintf("Failed to setup ping to %s", ip.String()), err)
		return false
	}
	pinger.Count = icmpProbeCount
	pinger.Interval = icmpProbeInterval
	err = pinger.Run()
	if err != nil {
		dest.Increment("connectivity.icmp.error", []string{})
		LogRouteError(route, fmt.Sprintf("Failed to ping %s", ip.String()), err)
		return false
	}

	stats := pinger.Statistics()
	if !icmpProbeSucceeded(int(stats.PacketsRecv), icmpProbeCount) {
		dest.Increment("connectivity.icmp.error", []string{})
		LogRouteError(route, fmt.Sprintf("Insufficient ping replies from %s (%d/%d received)", ip.String(), stats.PacketsRecv, icmpProbeCount), fmt.Errorf("packet loss %.0f%%", stats.PacketLoss))
		return false
	}

	Gauge("connectivity.icmp.packet_loss_pct", int(stats.PacketLoss), dest.tags())
	dest.Increment("connectivity.icmp.success", []string{})
	dest.Timer("connectivity.icmp", stats.AvgRtt, []string{})

	return true
}
