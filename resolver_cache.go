package main

import (
	"net"
	"sync"
	"time"
)

const (
	dnsCacheTTL         = 60 * time.Second
	dnsNegativeCacheTTL = 5 * time.Second
)

type dnsCacheEntry struct {
	ips    []net.IP
	err    error
	expiry time.Time
}

var dnsCache sync.Map

func lookupHostIPv4(host string) ([]net.IP, error) {
	if v, ok := dnsCache.Load(host); ok {
		entry := v.(dnsCacheEntry)
		if time.Now().Before(entry.expiry) {
			if entry.err != nil {
				return nil, entry.err
			}
			return cloneIPs(entry.ips), nil
		}
		dnsCache.Delete(host)
	}

	results, err := net.LookupIP(host)
	ttl := dnsCacheTTL
	if err != nil {
		ttl = dnsNegativeCacheTTL
	}

	ips := make([]net.IP, 0, len(results))
	for _, ip := range results {
		if ip.To4() != nil {
			ips = append(ips, ip)
		}
	}

	dnsCache.Store(host, dnsCacheEntry{
		ips:    cloneIPs(ips),
		err:    err,
		expiry: time.Now().Add(ttl),
	})

	if err != nil {
		return nil, err
	}
	return ips, nil
}

func cloneIPs(ips []net.IP) []net.IP {
	if len(ips) == 0 {
		return nil
	}
	out := make([]net.IP, len(ips))
	copy(out, ips)
	return out
}

func clearDNSCache() {
	dnsCache.Range(func(key, _ interface{}) bool {
		dnsCache.Delete(key)
		return true
	})
}
