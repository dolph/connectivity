package main

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

// TestStatsdConcurrentEnqueue exercises concurrent producers on the package-level
// statsd queue so go test -race observes the channel path used in production.
func TestStatsdConcurrentEnqueue(t *testing.T) {
	t.Cleanup(func() { drainQueue(t) })
	drainQueue(t)

	const producers = 32
	var wg sync.WaitGroup
	wg.Add(producers)
	for i := 0; i < producers; i++ {
		go func(n int) {
			defer wg.Done()
			Increment("connectivity.check", []string{fmt.Sprintf("worker:%d", n)})
		}(i)
	}
	wg.Wait()

	got := 0
	for {
		select {
		case <-queue:
			got++
		default:
			if got != producers {
				t.Fatalf("drained %d messages, want %d", got, producers)
			}
			return
		}
	}
}

// TestStatsdSenderConcurrentReceive runs StatsdSender against a local UDP sink
// while multiple goroutines enqueue metrics.
func TestStatsdSenderConcurrentReceive(t *testing.T) {
	t.Cleanup(func() { drainQueue(t) })
	drainQueue(t)

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	addr := conn.LocalAddr().(*net.UDPAddr)
	config := &Config{
		StatsdHost:     addr.IP.String(),
		StatsdPort:     addr.Port,
		StatsdProtocol: "udp",
	}
	go StatsdSender(config)

	const producers = 16
	var wg sync.WaitGroup
	wg.Add(producers)
	for i := 0; i < producers; i++ {
		go func() {
			defer wg.Done()
			Increment("connectivity.check", []string{"dest_host:example.com"})
		}()
	}
	wg.Wait()

	received := 0
	buf := make([]byte, 512)
	deadline := time.Now().Add(2 * time.Second)
	for received < producers && time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		if n > 0 {
			received++
		}
	}
	if received != producers {
		t.Fatalf("received %d UDP packets, want %d", received, producers)
	}
}
