package main

import (
	"testing"
)

// assertAtLeastOneIP ensures at least one result is returned
func assertAtLeastOneIP(t *testing.T, got *Source) {
	if len(got.IPs) == 0 {
		t.Errorf("len(GetLocalIPs().IPs) = %v; want >=1", len(got.IPs))
	}
}

// assertValidIPs ensures each result is a valid address
func assertValidIPs(t *testing.T, got *Source) {
	for idx, ip := range got.IPs {
		if ip == nil {
			t.Errorf("GetLocalIPs() = %v; want a valid representation of an IP address, got %s (idx=%d)", got, ip, idx)
		}
	}
}

// assertLoopbackOnlyReturnedByItself ensures loopback address is not returned
// unless it's returned by itself.
func assertLoopbackOnlyReturnedByItself(t *testing.T, got *Source) {
	// Either we get one non-loopback address, one loopback address, or all
	// non-loopback addresses... so we don't have to check the loopback
	// property if we only have one address.
	if len(got.IPs) != 1 {
		for idx, ip := range got.IPs {
			if ip.IsLoopback() {
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
