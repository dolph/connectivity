package main

import (
	"errors"
	"net/http"
)

// httpsClient does not follow redirects so a probe of the configured URL
// cannot be satisfied by an unrelated redirect target (see #15).
var httpsClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return errors.New("redirects disabled")
	},
}

// Performs a complete HTTP(S) request to the destination.
func HTTPS(dest *Destination) bool {
	dest.Increment("connectivity.http", []string{})
	_, err := httpsClient.Get(dest.URL)
	if err != nil {
		dest.Increment("connectivity.http.error", []string{})
		LogDestinationError(dest, "Failed HTTP GET", err)
		return false
	}
	dest.Increment("connectivity.http.success", []string{})
	return true
}
