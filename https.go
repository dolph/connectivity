package main

import (
	"log"
	"net/http"
)

func HTTPS(dest *Destination) bool {
	_, err := http.Get(dest.URL)
	if err != nil {
		log.Printf("%s failed: %v", dest, err)
		return false
	}
	return true
}
