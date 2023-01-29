package main

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
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
	if dest.Username != "" && dest.PasswordSet {
		return fmt.Sprintf("%s://%s:[...]@%s:%d%s", dest.Scheme, dest.Username, dest.Host, dest.Port, dest.Path)
	} else if dest.Username != "" && !dest.PasswordSet {
		return fmt.Sprintf("%s://%s@%s:%d%s", dest.Scheme, dest.Username, dest.Host, dest.Port, dest.Path)
	}
	return fmt.Sprintf("%s://%s:%d%s", dest.Scheme, dest.Host, dest.Port, dest.Path)
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
		Path:        url.Path}
}

func (dest *Destination) Monitor() {
	log.Printf("Monitoring connectivity to %s (%s)", dest, dest.Protocol)
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
