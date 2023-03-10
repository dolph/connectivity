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
	if r.SourceIP == nil && r.GatewayIP == nil {
		// If we failed to route through a gateway, log without a determined
		// source IP or outgoing iface. (Assume IPv4!)
		return fmt.Sprintf("[%s][%s › %s]", r.SourceHostname, GetLocalIPs(), r.DestinationIP)
	} else if r.SourceIP.String() == r.DestinationIP.String() || r.SourceIP.IsLoopback() {
		// If the source and destination are the same, the route is trivial.
		return fmt.Sprintf("[%s][%s][%s]", r.SourceHostname, r.SourceInterfaceName, r.DestinationIP)
	} else {
		return fmt.Sprintf("[%s][%s][%s][%s › %s » %s]", r.SourceHostname, r.SourceInterfaceName, r.SourceHardwareAddress, r.SourceIP, r.GatewayIP, r.DestinationIP)
	}
}

func GetRoute(ip net.IP) (*Route, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	route := &Route{
		SourceHostname:        hostname,
		SourceInterfaceName:   "",
		SourceHardwareAddress: nil,
		SourceIP:              nil,
		GatewayIP:             nil,
		DestinationIP:         ip}

	r, err := routing.New()
	if err != nil {
		return route, err
	}

	iface, gateway, source, err := r.Route(ip)
	if err != nil {
		// This is possibly a workaround until something like https://github.com/google/gopacket/pull/697 is released
		return route, err
	} else {
		route.SourceInterfaceName = iface.Name
		route.SourceHardwareAddress = iface.HardwareAddr
		route.SourceIP = source
		route.GatewayIP = gateway
		return route, nil
	}

}
