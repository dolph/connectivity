package main

import (
	"net"
	"os"
	"strings"
	"sync"
)

type Source struct {
	Hostname string
	IPs      []*net.IP
}

func NewSource() *Source {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	s := Source{
		Hostname: hostname}
	return &s
}

func (s *Source) String() string {
	IPs := []string{}
	for _, ip := range s.IPs {
		IPs = append(IPs, ip.String())
	}
	return "[" + s.Hostname + "][" + strings.Join(IPs, ",") + "]"
}

var (
	localSourceOnce sync.Once
	localSource     *Source
)

// GetLocalIPs returns the local source identity, discovered once per process.
func GetLocalIPs() *Source {
	localSourceOnce.Do(func() {
		localSource = discoverLocalIPs()
	})
	return localSource
}

// discoverLocalIPs lists non-loopback IPv4 addresses for this host.
func discoverLocalIPs() *Source {
	source := NewSource()
	addresses, err := net.InterfaceAddrs()
	if err == nil {
		for _, address := range addresses {
			// ignore loopback interfaces and IPv6 altogether
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				source.IPs = append(source.IPs, &ipnet.IP)
			}
		}
	}
	if len(source.IPs) == 0 {
		localhost := net.ParseIP("127.0.0.1")
		source.IPs = append(source.IPs, &localhost)
	}
	return source
}
