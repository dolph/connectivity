package main

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

/*

This module handles all the business logic of validating destinations by
orchestrating all the various connectivity verification steps that a given
destination might support.

*/

type Destination struct {
	Label       string
	URL         string
	Protocol    string
	Scheme      string
	Username    string
	Password    string
	PasswordSet bool
	Host        string
	Port        int
	Path        string
}

func (dest Destination) String() string {
	if dest.Label != "" {
		return fmt.Sprintf("%s:", dest.Label)
	} else {
		return fmt.Sprintf("%s:", dest.UrlString())
	}
}

func (dest *Destination) UrlString() string {
	s := fmt.Sprintf("%s://", dest.Scheme)

	// Suppress passwords if one is provided
	if dest.Username != "" && dest.PasswordSet {
		s += fmt.Sprintf("%s:[...]@", dest.Username)
	} else if dest.Username != "" && !dest.PasswordSet {
		s += fmt.Sprintf("%s@", dest.Username)
	}

	s += fmt.Sprintf("%s", dest.Host)

	// The port is -1 in the case of icmp:// where it's not relevant
	if dest.Port != -1 {
		s += fmt.Sprintf(":%d", dest.Port)
	}

	s += fmt.Sprintf("%s", dest.Path)

	if dest.Scheme != dest.Protocol {
		s += fmt.Sprintf(" (%s)", dest.Protocol)
	}

	return s
}

func (dest *Destination) tags() []string {
	return []string{
		fmt.Sprintf("dest_label:%s", dest.Label),
		fmt.Sprintf("dest_scheme:%s", dest.Scheme),
		fmt.Sprintf("dest_host:%s", dest.Host),
		fmt.Sprintf("dest_port:%d", dest.Port),
	}
}

func (dest *Destination) Increment(metric string, tags []string) {
	tags = append(tags, dest.tags()...)
	Increment(metric, tags)
}

func (dest *Destination) Timer(metric string, took time.Duration) {
	Timer(metric, took, dest.tags())
}

func NewDestination(u Url) (*Destination, error) {
	url, err := url.Parse(u.Url)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v: Failed to parse URL: %v", u, err))
	}

	// Determine host
	host := url.Hostname()
	if host == "" {
		return nil, errors.New(fmt.Sprintf("%v: Failed to parse a host in URL: %v", u, u.Url))
	}

	// Determine scheme
	scheme := strings.ToLower(url.Scheme)

	// Determine port number
	port := url.Port()
	if port != "" && scheme == "icmp" {
		return nil, errors.New(fmt.Sprintf("%s: ICMP cannot be used with a port number: %v", u, u.Url))
	}
	var portNumber int
	if port != "" {
		portNumber, err = strconv.Atoi(url.Port())
	}
	if port == "" || err != nil {
		portNumber, err = net.LookupPort("tcp", scheme)

		if err != nil {
			// Custom ports for schemes unknown to Go can be set here to avoid erroring
			// If Go adopts support for one of these, this code won't be reached.
			if scheme == "nats" {
				portNumber = 4222
			} else if scheme == "icmp" {
				portNumber = -1
			} else {
				return nil, errors.New(fmt.Sprintf("%s: Unsupported scheme (try specifying tcp:// or udp:// and an explicit port, or icmp:// for ping-only): %s", u, err))
			}
		}
	}

	// Determine protocol
	protocol := "tcp"
	if scheme == "udp" || scheme == "icmp" {
		protocol = scheme
	}

	username := url.User.Username()
	password, passwordSet := url.User.Password()

	return &Destination{
			Label:       u.Label,
			URL:         u.Url,
			Protocol:    protocol,
			Scheme:      scheme,
			Username:    username,
			Password:    password,
			PasswordSet: passwordSet,
			Host:        host,
			Port:        portNumber,
			Path:        url.Path},
		nil
}

func (dest *Destination) Check() bool {
	// Assume the destination is reachable until proven otherwise
	reachable := true

	dnsResults, err := Lookup(dest)
	if err != nil {
		LogDestinationError(dest, "Failed to resolve host", err)
		reachable = false
	}

	if reachable {
		for _, ip := range dnsResults {
			// Check that this isn't an IPv6 result
			if !strings.Contains(ip.String(), ":") {
				// Check destination IP for routability
				route, err := GetRoute(ip)
				if err != nil {
					LogDestinationError(dest, fmt.Sprintf("Failed to route to %s", ip.String()), err)
					return false
				}

				if dest.Protocol == "icmp" {
					reachable = reachable && Ping(route, dest, ip)
				} else {
					reachable = reachable && Dial(route, dest, ip)
				}
			}
		}
	}

	if reachable {
		if dest.Scheme == "http" || dest.Scheme == "https" {
			reachable = reachable && HTTPS(dest)
		}
	}

	return reachable
}

func (dest *Destination) Monitor() {
	confidence := 1

	for {
		dest.Increment("connectivity.check", []string{})
		reachable := dest.Check()

		if reachable {
			confidence += 1
			if confidence > 10 {
				confidence = 10
			}
		} else {
			dest.Increment("connectivity.check.error", []string{})
			confidence -= 1
			if confidence < 1 {
				confidence = 1
			}
		}

		time.Sleep(time.Duration(confidence) * time.Minute)
	}
}

func (dest *Destination) WaitFor() {
	for {
		dest.Increment("connectivity.check", []string{})
		reachable := dest.Check()

		if reachable {
			LogDestination(dest, "Validated")
			return
		} else {
			dest.Increment("connectivity.check.error", []string{})
		}

		time.Sleep(15 * time.Second)
	}
}
