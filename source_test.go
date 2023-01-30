package main

import (
	"net"
	"testing"
)

// assertAtLeastOneIP ensures at least one result is returned
func assertAtLeastOneIP(t *testing.T, got []string) {
	if len(got) == 0 {
		t.Errorf("len(GetLocalIPs()) = %v; want >=1", len(got))
	}
}

// assertValidIPs ensures each result is a valid address
func assertValidIPs(t *testing.T, got []string) {
	for idx, ip := range GetLocalIPs() {
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			t.Errorf("GetLocalIPs() = %v; want a valid textual representation of an IP address, not %s (idx=%d)", got, parsedIP, idx)
		}
	}
}

// assertLoopbackOnlyReturnedByItself ensures loopback address is not returned
// unless it's returned by itself.
func assertLoopbackOnlyReturnedByItself(t *testing.T, got []string) {
	// Either we get one non-loopback address, one loopback address, or all
	// non-loopback addresses... so we don't have to check the loopback
	// property if we only have one address.
	if len(got) != 1 {
		for idx, ip := range got {
			if net.ParseIP(ip).IsLoopback() {
				t.Errorf("GetLocalIPs() = %v; want either one loopback address or all non-loopback addresses, not %s (idx=%d)", got, ip, idx)
			}
		}
	}
}

func TestGetLocalIPs(t *testing.T) {
	got := GetLocalIPs()
	assertAtLeastOneIP(t, got)
	assertValidIPs(t, got)
	assertLoopbackOnlyReturnedByItself(t, got)
}
