package main

import (
	"net/http"
)

// Performs a complete HTTP(S) request to the destination.
func HTTPS(dest *Destination) bool {
	dest.Increment("connectivity.http", []string{})
	_, err := http.Get(dest.URL)
	if err != nil {
		dest.Increment("connectivity.http.error", []string{})
		LogDestinationError(dest, "Failed HTTP GET", err)
		return false
	}
	dest.Increment("connectivity.http.success", []string{})
	return true
}
