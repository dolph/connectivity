package main

import (
	"strings"
	"testing"
)

func FuzzNewDestination(f *testing.F) {
	f.Add("https://example.com")
	f.Add("icmp://example.com")
	f.Add("tcp://127.0.0.1:443")
	f.Add("not a url")

	f.Fuzz(func(t *testing.T, raw string) {
		if strings.Contains(raw, "\x00") {
			return
		}
		_, _ = NewDestination(Url{Url: raw})
	})
}
