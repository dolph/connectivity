package main

import "testing"

func FuzzParseConfigYAML(f *testing.F) {
	f.Add([]byte("api: https://example.com/health\n"))
	f.Add([]byte("{}\n"))
	f.Add([]byte("not: [valid, yaml"))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseConfigYAML(data)
	})
}
