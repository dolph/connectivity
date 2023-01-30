package main

import (
	"testing"
)

func assertSchemeEquals(t *testing.T, got *Destination, want string) {
	if got.Scheme != want {
		t.Errorf("dest.Scheme = %s; want %s", got, want)
	}
}

func assertHostEquals(t *testing.T, got *Destination, want string) {
	if got.Host != want {
		t.Errorf("dest.Host = %s; want %s", got, want)
	}
}

func assertPortEquals(t *testing.T, got *Destination, want int) {
	if got.Port != want {
		t.Errorf("dest.Port = %s; want %d", got, want)
	}
}

func TestMinimalHttpUrl(t *testing.T) {
	got := NewDestination("http://host")
	assertSchemeEquals(t, got, "http")
	assertHostEquals(t, got, "host")
	assertPortEquals(t, got, 80)
}

func TestMinimalHttpsUrl(t *testing.T) {
	got := NewDestination("https://host")
	assertSchemeEquals(t, got, "https")
	assertHostEquals(t, got, "host")
	assertPortEquals(t, got, 443)
}

func TestSchemeNormalization(t *testing.T) {
	got := NewDestination("HTtP://host")
	assertSchemeEquals(t, got, "http")
	assertHostEquals(t, got, "host")
	assertPortEquals(t, got, 80)
}
