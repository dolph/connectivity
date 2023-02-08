package main

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Destination struct {
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

func (dest *Destination) String() string {
	port := ""
	if dest.Port != -1 {
		port = fmt.Sprintf(":%d", dest.Port)
	}

	if dest.Username != "" && dest.PasswordSet {
		return fmt.Sprintf("%s://%s:[...]@%s%s%s (%s)", dest.Scheme, dest.Username, dest.Host, port, dest.Path, dest.Protocol)
	} else if dest.Username != "" && !dest.PasswordSet {
		return fmt.Sprintf("%s://%s@%s%s%s (%s)", dest.Scheme, dest.Username, dest.Host, port, dest.Path, dest.Protocol)
	}
	return fmt.Sprintf("%s://%s%s%s (%s)", dest.Scheme, dest.Host, port, dest.Path, dest.Protocol)
}

func (dest *Destination) tags() []string {
	return []string{
		fmt.Sprintf("dest_url:%s", dest.String()),
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

func NewDestination(dest string) (*Destination, error) {
	url, err := url.Parse(dest)
	if err != nil {
		log.Printf("Failed to parse URL (%s): %s", url, err)
		return nil, err
	}

	// Determine host
	host := url.Hostname()
	if host == "" {
		log.Printf("Failed to parse a host in URL (%s): %s", url, err)
		return nil, err
	}

	// Determine scheme
	scheme := strings.ToLower(url.Scheme)

	// Determine port number
	port := url.Port()
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
				log.Printf("Unsupported scheme (try specifying tcp:// or udp:// and an explicit port, or icmp:// for ping-only) (%s): %s", url, err)
				return nil, err
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
			URL:         dest,
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
		log.Printf("%s Failed to resolve %s (%v)", GetLocalIPs(), dest.Host, err)
		reachable = false
	}

	if reachable {
		for _, ip := range dnsResults {
			// Check that this isn't an IPv6 result
			if !strings.Contains(ip.String(), ":") {
				if dest.Protocol == "icmp" {
					reachable = reachable && Ping(dest, ip)
				} else {
					reachable = reachable && Dial(dest, ip)
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
	log.Printf("Monitoring connectivity to %s (%s)", dest, dest.Protocol)

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
	log.Printf("Waiting for connectivity to %s (%s)", dest, dest.Protocol)

	for {
		dest.Increment("connectivity.check", []string{})
		reachable := dest.Check()

		if reachable {
			log.Printf("Validated %s (%s)", dest, dest.Protocol)
			return
		} else {
			dest.Increment("connectivity.check.error", []string{})
		}

		time.Sleep(15 * time.Second)
	}
}
