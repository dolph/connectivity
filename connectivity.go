package main

import (
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Basic design:
	// Handoff each URL to a monitoring process that will be in charge of it.
	// Each URL is pre-processed to understand how it can be validated.
	// 1. Can it be parsed into a valid URL?
	// 2. Can it be resolved?
	// 3. Can the host be pinged or each resolved host:port dialed?
	// 4. If it can be reached, handoff to a protocol-specific handler that validates application-level connectivity.
	// 5. Validate the expected outcome. A "permission denied" may be expected from SSH, for example.

	// Validate all destinations before beginning any monitoring
	var destinations []*Destination
	for idx, url := range os.Args {
		if idx != 0 {
			destinations = append(destinations, NewDestination(url))
		}
	}

	for _, dest := range destinations {
		go dest.Monitor()
	}

	// Sleep forever
	select {}
}

type Destination struct {
	URL      string
	Protocol string
	Scheme   string
	Host     string
	Port     int
}

func (dest Destination) String() string {
	return dest.URL
}

func NewDestination(dest string) *Destination {
	url, err := url.Parse(dest)
	if err != nil {
		log.Fatalf("Failed to parse URL (%s): %s", url, err)
		os.Exit(2)
	}

	// Determine host
	host := url.Hostname()
	if host == "" {
		log.Fatalf("Failed to parse a host in URL (%s): %s", url, err)
		os.Exit(2)
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
			log.Fatalf("Unsupported scheme (try specifying tcp:// or udp:// and an explicit port) (%s): %s", url, err)
			os.Exit(2)
		}
	}

	// Determine protocol
	protocol := "tcp"
	if scheme == "udp" {
		protocol = scheme
	}

	return &Destination{
		URL:      dest,
		Protocol: protocol,
		Scheme:   scheme,
		Host:     host,
		Port:     portNumber}
}

func (dest *Destination) Monitor() {
	log.Printf("Monitoring connectivity to %s (scheme=%s host=%s port=%d) (%s)", dest, dest.Scheme, dest.Host, dest.Port, dest.Protocol)
	confidence := 1

	for {
		// Assume the destination is unreachable until proven otherwise
		reachable := false

		for _, ip := range Lookup(dest) {
			if !strings.Contains(ip, ":") {
				reachable = reachable || Dial(dest, ip)
			}
		}

		if reachable {
			if dest.Scheme == "http" || dest.Scheme == "https" {
				HTTPS(dest)
			}
		}

		if reachable {
			confidence += 1
			if confidence > 10 {
				confidence = 10
			}
		} else {
			confidence -= 1
			if confidence < 1 {
				confidence = 1
			}
		}
		time.Sleep(time.Duration(confidence) * time.Minute)
	}
}
