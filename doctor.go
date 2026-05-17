package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Doctor runs each connectivity layer with verbose logging for incident triage.
func (dest *Destination) Doctor() bool {
	LogDestination(dest, "doctor: starting diagnostic")

	reachable := true

	dnsResults, err := Lookup(dest)
	if err != nil {
		LogDestinationError(dest, "doctor: DNS lookup failed", err)
		return false
	}

	if len(dnsResults) == 0 {
		LogDestination(dest, "doctor: DNS returned no addresses")
		return false
	}

	for _, ip := range dnsResults {
		if strings.Contains(ip.String(), ":") {
			LogDestination(dest, "doctor: skipping IPv6 address "+ip.String())
			continue
		}

		LogDestination(dest, "doctor: DNS resolved "+ip.String())

		route, routeErr := GetRoute(ip)
		if routeErr != nil {
			LogDestination(dest, fmt.Sprintf("doctor: route lookup unavailable for %s (continuing): %v", ip.String(), routeErr))
		} else {
			LogRoute(route, "doctor: route for "+ip.String())
		}

		var layerOK bool
		if dest.Protocol == "icmp" {
			layerOK = Ping(route, dest, ip)
		} else {
			layerOK = Dial(route, dest, ip)
		}
		if !layerOK {
			reachable = false
		}
	}

	if dest.Scheme == "http" || dest.Scheme == "https" {
		if !doctorHTTPS(dest) {
			reachable = false
		}
	}

	if reachable {
		LogDestination(dest, "doctor: all layers passed")
	} else {
		LogDestination(dest, "doctor: one or more layers failed")
	}

	return reachable
}

func doctorHTTPS(dest *Destination) bool {
	LogDestination(dest, "doctor: HTTP GET "+dest.URL)

	resp, err := http.Get(dest.URL)
	if err != nil {
		LogDestinationError(dest, "doctor: HTTP GET failed", err)
		return false
	}
	defer resp.Body.Close()

	LogDestination(dest, fmt.Sprintf("doctor: HTTP status %s", resp.Status))

	snippet, readErr := io.ReadAll(io.LimitReader(resp.Body, 256))
	if readErr != nil {
		LogDestination(dest, fmt.Sprintf("doctor: could not read response body: %v", readErr))
	} else if len(snippet) > 0 {
		LogDestination(dest, fmt.Sprintf("doctor: response body (first %d bytes): %q", len(snippet), snippet))
	} else {
		LogDestination(dest, "doctor: response body empty")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		LogDestination(dest, fmt.Sprintf("doctor: unexpected HTTP status %d", resp.StatusCode))
		return false
	}

	return true
}

func DoctorLoop(destinations []*Destination) bool {
	ok := true
	for _, dest := range destinations {
		if !dest.Doctor() {
			ok = false
		}
	}
	return ok
}
