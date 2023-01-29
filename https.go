package main

import (
	"log"
	"net/http"
)

// Performs a complete HTTP(S) request to the destination.
func HTTPS(dest *Destination) bool {
	_, err := http.Get(dest.URL)
	if err != nil {
		log.Printf("%s Failed HTTP GET %s: %v", GetLocalIPs(), dest, err)
		return false
	}
	return true
}
