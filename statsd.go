package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

var queue = make(chan string, 100)

func Increment(metric string, tags []string) {
	Count(metric, 1, tags)
}

func Count(metric string, value int, tags []string) {
	queue <- fmt.Sprintf("%s:%d|c|#%s", metric, value, formatTags(tags))
}

func Timer(metric string, took time.Duration, tags []string) {
	queue <- fmt.Sprintf("%s:%d|ms|#%s", metric, took/1e6, formatTags(tags))
}

func Gauge(metric string, value int, tags []string) {
	queue <- fmt.Sprintf("%s:%d|g|#%s", metric, value, formatTags(tags))
}

func formatTags(tags []string) string {
	return strings.Join(tags[:], ",")
}

func StatsdSender(config *Config) {
	for s := range queue {
		statsdHostPort := fmt.Sprintf("%s:%d", config.StatsdHost, config.StatsdPort)
		if conn, err := net.Dial(config.StatsdProtocol, statsdHostPort); err == nil {
			io.WriteString(conn, s)
			conn.Close()
		}
	}
}
