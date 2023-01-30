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

func assertNoError(t *testing.T, got *Destination, err error) {
	if err != nil {
		t.Errorf("got = %s with %v; want no err", got, err)
	}
}

func assertError(t *testing.T, got *Destination, err error) {
	if err == nil {
		t.Errorf("got = %s; want err", got)
	}
}

func TestMinimalHttpUrl(t *testing.T) {
	got, err := NewDestination("http://host")
	assertNoError(t, got, err)
	assertSchemeEquals(t, got, "http")
	assertHostEquals(t, got, "host")
	assertPortEquals(t, got, 80)
}

func TestMinimalHttpsUrl(t *testing.T) {
	got, err := NewDestination("https://host")
	assertNoError(t, got, err)
	assertSchemeEquals(t, got, "https")
	assertHostEquals(t, got, "host")
	assertPortEquals(t, got, 443)
}

func TestMinimalMysqlUrl(t *testing.T) {
	got, err := NewDestination("mysql://host")
	assertNoError(t, got, err)
	assertSchemeEquals(t, got, "mysql")
	assertHostEquals(t, got, "host")
	assertPortEquals(t, got, 3306)
}

func TestMinimalPostgresUrl(t *testing.T) {
	got, err := NewDestination("postgres://host")
	assertNoError(t, got, err)
	assertSchemeEquals(t, got, "postgres")
	assertHostEquals(t, got, "host")
	assertPortEquals(t, got, 5432)
}

func TestMinimalNatsUrl(t *testing.T) {
	got, err := NewDestination("nats://host")
	assertNoError(t, got, err)
	assertSchemeEquals(t, got, "nats")
	assertHostEquals(t, got, "host")
	assertPortEquals(t, got, 4222)
}

func TestSchemeNormalization(t *testing.T) {
	got, err := NewDestination("HTtP://host")
	assertNoError(t, got, err)
	assertSchemeEquals(t, got, "http")
	assertHostEquals(t, got, "host")
	assertPortEquals(t, got, 80)
}

func TestTcpUrlWithoutPort(t *testing.T) {
	got, err := NewDestination("tcp://host")
	assertError(t, got, err)
}

func TestTcpUrlWithPort(t *testing.T) {
	got, err := NewDestination("tcp://host:123")
	assertNoError(t, got, err)
	assertSchemeEquals(t, got, "tcp")
	assertHostEquals(t, got, "host")
	assertPortEquals(t, got, 123)
}

func TestUdpUrlWithoutPort(t *testing.T) {
	got, err := NewDestination("udp://host")
	assertError(t, got, err)
}

func TestUdpUrlWithPort(t *testing.T) {
	got, err := NewDestination("udp://host:123")
	assertNoError(t, got, err)
	assertSchemeEquals(t, got, "udp")
	assertHostEquals(t, got, "host")
	assertPortEquals(t, got, 123)
}
