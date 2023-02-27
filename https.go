package main

import (
	"net/http"
)

// Performs a complete HTTP(S) request to the destination.
func HTTPS(dest *Destination) bool {
	_, err := http.Get(dest.URL)
	if err != nil {
		LogDestinationError(dest, "Failed HTTP GET", err)
		return false
	}
	return true
}
