package main

import (
	"errors"
	"net"
	"os"
	"sync"

	"github.com/google/gopacket/routing"
)

var errRouteTableUnavailable = errors.New("route table unavailable")

var (
	routeRouterMu sync.RWMutex
	routeRouter   routing.Router
	routeHostname string
)

func init() {
	routeHostname, _ = os.Hostname()
	if routeHostname == "" {
		routeHostname = "localhost"
	}
}

func refreshRouteRouter() {
	var r routing.Router
	var err error
	func() {
		defer func() {
			if recover() != nil {
				err = errRouteTableUnavailable
			}
		}()
		r, err = routing.New()
	}()
	routeRouterMu.Lock()
	defer routeRouterMu.Unlock()
	if err == nil && r != nil {
		routeRouter = r
	}
}

func lookupRoute(ip net.IP) (*net.Interface, net.IP, net.IP, error) {
	routeRouterMu.RLock()
	r := routeRouter
	routeRouterMu.RUnlock()

	if r == nil {
		refreshRouteRouter()
		routeRouterMu.RLock()
		r = routeRouter
		routeRouterMu.RUnlock()
	}

	if r == nil {
		return nil, nil, nil, errRouteTableUnavailable
	}

	iface, gateway, source, err := r.Route(ip)
	if err == nil {
		return iface, gateway, source, nil
	}

	refreshRouteRouter()
	routeRouterMu.RLock()
	r = routeRouter
	routeRouterMu.RUnlock()
	if r == nil {
		return nil, nil, nil, err
	}
	return r.Route(ip)
}

func clearRouteRouterForTest() {
	routeRouterMu.Lock()
	routeRouter = nil
	routeRouterMu.Unlock()
	refreshRouteRouter()
}
