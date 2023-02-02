package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
)

func assertRouteSourceIPEquals(t *testing.T, got *Route, want string) {
	if got.SourceIP.String() != want {
		t.Errorf("Route.SourceIP = %s; want %s", got.SourceIP.String(), want)
	}
}

func assertRouteGatewayIPEquals(t *testing.T, got *Route, want string) {
	if got.GatewayIP.String() != want {
		t.Errorf("Route.GatewayIP = %s; want %s", got.GatewayIP.String(), want)
	}
}

func assertRouteDestinationIPEquals(t *testing.T, got *Route, want string) {
	if got.DestinationIP.String() != want {
		t.Errorf("Route.DestinationIP = %s; want %s", got.DestinationIP.String(), want)
	}
}

func assertRouteStringEquals(t *testing.T, got *Route, want string) {
	if got.String() != want {
		t.Errorf("Route.String() = %s; want %s", got.String(), want)
	}
}
func assertRoutable(t *testing.T, got *Route) {
	if strings.Count(got.String(), "->") != 2 {
		t.Errorf("Route.String() = %s; want a hop before destination", got.String())
	}
	if got.SourceIP.String() == got.GatewayIP.String() {
		t.Errorf("Route.SourceIP.String() = Route.GatewayIP.String() = %s ; want different a gateway", got.SourceIP.String())
	}
	if got.SourceIP.String() == got.DestinationIP.String() {
		t.Errorf("Route.SourceIP.String() = Route.DestinationIP.String() = %s ; want different a destination", got.SourceIP.String())
	}
	if got.GatewayIP.String() == got.DestinationIP.String() {
		t.Errorf("Route.GatewayIP.String() = Route.DestinationIP.String() = %s ; want different a gateway", got.GatewayIP.String())
	}
}
func assertNil(t *testing.T, err error) {
	if err != nil {
		t.Errorf("err = %s; want nil", err)
	}
}

func TestRouteToLoopback1(t *testing.T) {
	route, err := GetRoute(net.ParseIP("127.0.0.1"))
	assertNil(t, err)
	assertRouteSourceIPEquals(t, route, "127.0.0.1")
	assertRouteGatewayIPEquals(t, route, "<nil>")
	assertRouteDestinationIPEquals(t, route, "127.0.0.1")
	hostname, _ := os.Hostname()
	assertRouteStringEquals(t, route, fmt.Sprintf("[%s][lo][127.0.0.1]", hostname))
}

func TestRouteToLoopback2(t *testing.T) {
	route, err := GetRoute(net.ParseIP("127.0.1.1"))
	assertNil(t, err)
	assertRouteSourceIPEquals(t, route, "127.0.0.1")
	assertRouteGatewayIPEquals(t, route, "<nil>")
	assertRouteDestinationIPEquals(t, route, "127.0.1.1")
	hostname, _ := os.Hostname()
	assertRouteStringEquals(t, route, fmt.Sprintf("[%s][lo][127.0.1.1]", hostname))
}

func TestRouteToLoopback3(t *testing.T) {
	route, err := GetRoute(net.ParseIP("127.1.0.1"))
	assertNil(t, err)
	assertRouteSourceIPEquals(t, route, "127.0.0.1")
	assertRouteGatewayIPEquals(t, route, "<nil>")
	assertRouteDestinationIPEquals(t, route, "127.1.0.1")
	hostname, _ := os.Hostname()
	assertRouteStringEquals(t, route, fmt.Sprintf("[%s][lo][127.1.0.1]", hostname))
}

func TestRouteToPublic(t *testing.T) {
	route, err := GetRoute(net.ParseIP("1.2.3.4"))
	assertNil(t, err)
	assertRouteDestinationIPEquals(t, route, "1.2.3.4")
	assertRoutable(t, route)
}
