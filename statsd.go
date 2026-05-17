package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

var queue = make(chan string, 100)

func enqueueMetric(line string) {
	select {
	case queue <- line:
	default:
		// Drop when statsd is slow or down so checks never block on metrics.
	}
}

func Increment(metric string, tags []string) {
	Count(metric, 1, tags)
}

func Count(metric string, value int, tags []string) {
	enqueueMetric(fmt.Sprintf("%s:%d|c|#%s", metric, value, formatTags(tags)))
}

func Timer(metric string, took time.Duration, tags []string) {
	enqueueMetric(fmt.Sprintf("%s:%d|ms|#%s", metric, took/1e6, formatTags(tags)))
}

func Gauge(metric string, value int, tags []string) {
	enqueueMetric(fmt.Sprintf("%s:%d|g|#%s", metric, value, formatTags(tags)))
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
	s = strings.Replace(s, "\n", "-", -1)
	s = strings.Replace(s, "\r", "-", -1)
	return s
}

func StatsdSender(config *Config) {
	addr := fmt.Sprintf("%s:%d", config.StatsdHost, config.StatsdPort)
	var conn net.Conn
	var logMu sync.Mutex
	var lastErrLog time.Time

	logFailure := func(msg string, err error) {
		logMu.Lock()
		defer logMu.Unlock()
		if time.Since(lastErrLog) < 10*time.Second {
			return
		}
		lastErrLog = time.Now()
		log.Printf("%s: %v", msg, err)
	}

	for s := range queue {
		if conn == nil {
			c, err := net.Dial(config.StatsdProtocol, addr)
			if err != nil {
				logFailure("statsd dial failed", err)
				continue
			}
			conn = c
		}

		if _, err := io.WriteString(conn, s); err != nil {
			logFailure("statsd write failed", err)
			_ = conn.Close()
			conn = nil
		}
	}

	if conn != nil {
		_ = conn.Close()
	}
}
