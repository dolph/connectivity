package main

import (
	"fmt"
	"net"
	"os"

	"github.com/google/gopacket/routing"
)

type Route struct {
	SourceHostname        string
	SourceInterfaceName   string
	SourceHardwareAddress net.HardwareAddr
	SourceIP              net.IP
	GatewayIP             net.IP
	DestinationIP         net.IP
}

func (r *Route) String() string {
	if r.SourceIP.String() == r.DestinationIP.String() || r.SourceIP.IsLoopback() {
		// If the source and destination are the same, the route is trivial.
		return fmt.Sprintf("[%s][%s][%s]", r.SourceHostname, r.SourceInterfaceName, r.DestinationIP)
	} else {
		return fmt.Sprintf("[%s][%s][%s][%s › %s » %s]", r.SourceHostname, r.SourceInterfaceName, r.SourceHardwareAddress, r.SourceIP, r.GatewayIP, r.DestinationIP)
	}
}

func GetRoute(ip net.IP) (*Route, error) {
	r, err := routing.New()
	if err != nil {
		return nil, err
	}

	iface, gateway, source, err := r.Route(ip)
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}

	return &Route{
			SourceHostname:        hostname,
			SourceInterfaceName:   iface.Name,
			SourceHardwareAddress: iface.HardwareAddr,
			SourceIP:              source,
			GatewayIP:             gateway,
			DestinationIP:         ip},
		nil
}
