package main

import (
	"log"
	"net"

	probing "github.com/prometheus-community/pro-bing"
)

func Ping(dest *Destination, ip net.IP) bool {
	pinger, err := probing.NewPinger(ip.String())
	if err != nil {
		log.Printf("Failed to setup ping to %s (%v)", ip.String(), err)
		return false
	}
	pinger.Count = 1
	err = pinger.Run()
	if err != nil {
		log.Printf("Failed to ping %s (%v)", ip.String(), err)
		return false
	}

	// TODO: emit stats to statsd. have to keep the pinger around to get stddev
	// stats := pinger.Statistics()
	// stats.AvgRtt.Milliseconds()

	return true
}
