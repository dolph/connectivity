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

func EscapeTag(s string) string {
	// Replace all special characters used in the statsd wire protocol
	s = strings.Replace(s, ":", "-", -1)
	s = strings.Replace(s, "|", "-", -1)
	s = strings.Replace(s, ",", "-", -1)
	s = strings.Replace(s, "@", "-", -1)
	return s
}

func emitStatsd(config *Config, payload string) {
	statsdHostPort := fmt.Sprintf("%s:%d", config.StatsdHost, config.StatsdPort)
	conn, err := net.Dial(config.StatsdProtocol, statsdHostPort)
	if err != nil {
		return
	}
	_, _ = io.WriteString(conn, payload)
	_ = conn.Close()
}

func StatsdSender(config *Config) {
	for s := range queue {
		emitStatsd(config, s)
	}
}

// DrainStatsd flushes queued metrics for up to timeout before process exit.
func DrainStatsd(config *Config, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case s := <-queue:
			emitStatsd(config, s)
		default:
			return
		}
	}
}
