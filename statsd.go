package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// queue is the package-level statsd message buffer that main wires
// StatsdSender to consume. The lowercase helpers (count, timer, gauge,
// increment, statsdSender) take an explicit queue parameter so tests can
// drive the producer/consumer surface with a hermetic channel; see #49 and
// concurrency_test.go.
//
// The capacity-100 default and blocking send are the back-pressure hazard
// described in #11. The seam below does NOT fix that bug — it just makes it
// observable from a test. See #17 for the broader process-lifecycle work
// that needs to land before non-blocking sends and drain-on-shutdown can be
// done correctly.
var queue = make(chan string, 100)

func Increment(metric string, tags []string) {
	increment(queue, metric, tags)
}

func Count(metric string, value int, tags []string) {
	count(queue, metric, value, tags)
}

func Timer(metric string, took time.Duration, tags []string) {
	timer(queue, metric, took, tags)
}

func Gauge(metric string, value int, tags []string) {
	gauge(queue, metric, value, tags)
}

func increment(q chan<- string, metric string, tags []string) {
	count(q, metric, 1, tags)
}

func count(q chan<- string, metric string, value int, tags []string) {
	q <- fmt.Sprintf("%s:%d|c|#%s", metric, value, formatTags(tags))
}

func timer(q chan<- string, metric string, took time.Duration, tags []string) {
	q <- fmt.Sprintf("%s:%d|ms|#%s", metric, took/1e6, formatTags(tags))
}

func gauge(q chan<- string, metric string, value int, tags []string) {
	q <- fmt.Sprintf("%s:%d|g|#%s", metric, value, formatTags(tags))
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

func StatsdSender(config *Config) {
	statsdSender(config, queue)
}

func statsdSender(config *Config, q <-chan string) {
	for s := range q {
		statsdHostPort := fmt.Sprintf("%s:%d", config.StatsdHost, config.StatsdPort)
		if conn, err := net.Dial(config.StatsdProtocol, statsdHostPort); err == nil {
			io.WriteString(conn, s)
			conn.Close()
		}
	}
}
