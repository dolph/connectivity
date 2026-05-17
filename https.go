package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const httpsCheckTimeout = 30 * time.Second

var httpsClient = &http.Client{
	Timeout: httpsCheckTimeout,
}

// Performs a complete HTTP(S) request to the destination.
func HTTPS(dest *Destination) bool {
	dest.Increment("connectivity.http", []string{})
	resp, err := httpsClient.Get(dest.URL)
	if err != nil {
		dest.Increment("connectivity.http.error", []string{})
		LogDestinationError(dest, "Failed HTTP GET", err)
		return false
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		dest.Increment("connectivity.http.error", []string{})
		LogDestinationError(dest, "Unexpected HTTP status", fmt.Errorf("HTTP %d", resp.StatusCode))
		return false
	}

	dest.Increment("connectivity.http.success", []string{})
	return true
}
