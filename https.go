package main

import (
	"net/http"
	"strings"
)

func httpCheckErrorMessage(err error) string {
	if err == nil {
		return "HTTP request failed"
	}
	msg := err.Error()
	if strings.Contains(msg, "x509:") || strings.HasPrefix(msg, "tls:") {
		return "TLS handshake failed"
	}
	return "HTTP request failed"
}

// Performs a complete HTTP(S) request to the destination.
func HTTPS(dest *Destination) bool {
	dest.Increment("connectivity.http", []string{})
	_, err := http.Get(dest.URL)
	if err != nil {
		dest.Increment("connectivity.http.error", []string{})
		LogDestinationError(dest, httpCheckErrorMessage(err), err)
		return false
	}
	dest.Increment("connectivity.http.success", []string{})
	return true
}
