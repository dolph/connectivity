//go:build linux

package main

import (
	"net"

	"github.com/google/gopacket/routing"
)

func lookupRoute(ip net.IP, route *Route) error {
	r, err := routing.New()
	if err != nil {
		return err
	}

	iface, gateway, source, err := r.Route(ip)
	if err != nil {
		return err
	}

	route.SourceInterfaceName = iface.Name
	route.SourceHardwareAddress = iface.HardwareAddr
	route.SourceIP = source
	route.GatewayIP = gateway
	return nil
}
