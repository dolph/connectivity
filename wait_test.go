package main

import (
	"runtime"
	"testing"
	"time"
)

func TestWaitForUntil_TimesOut(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("GetRoute is only implemented on Linux")
	}

	t.Cleanup(func() { drainQueue(t) })
	drainQueue(t)

	dest := &Destination{
		Label:    "unreachable",
		URL:      "http://127.0.0.1:1/",
		Protocol: "tcp",
		Scheme:   "http",
		Host:     "127.0.0.1",
		Port:     1,
	}

	deadline := time.Now().Add(50 * time.Millisecond)
	start := time.Now()
	if dest.WaitForUntil(deadline) {
		t.Fatal("WaitForUntil() = true; want false on timeout")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("WaitForUntil took %v; expected to return near deadline", elapsed)
	}
}
