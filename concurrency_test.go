package main

import (
	"net"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

// This file holds hermetic tests for the concurrent surfaces of the package:
// the statsd producer/consumer goroutines, WaitLoop's per-destination
// goroutines, and the Monitor confidence loop. Each test exercises a seam
// (lowercase variant of the public function) so it can drive the concurrent
// code path synchronously without spinning up production goroutines.
//
// Refs #49 — these tests are the reason -race in CI is no longer a false
// signal: they actually execute the concurrent producer/consumer surface.

// TestStatsdCountBlocksWhenQueueFull pins the back-pressure hazard documented
// in #11: when StatsdSender cannot drain (statsd down, slow TCP, etc.) the
// 100-slot queue fills up and every metric helper blocks. The test uses an
// injected queue with capacity 1 so it can be saturated cheaply.
//
// Refs #11 -- flip when fixed: once the producers become non-blocking
// (select-with-default + drop_and_count_drops), this test should be updated
// to assert the drop counter increments and Count returns promptly.
func TestStatsdCountBlocksWhenQueueFull(t *testing.T) {
	q := make(chan string, 1)
	// Saturate the queue. No consumer is running, so the next send will block.
	count(q, "preload", 1, []string{"t:v"})

	done := make(chan struct{})
	go func() {
		count(q, "should-block", 1, []string{"t:v"})
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("count(q, ...) returned while queue was full; want it to block " +
			"(this is the bug from #11; flip the assertion when #11 is fixed)")
	case <-time.After(50 * time.Millisecond):
		// Expected: the send is blocked. Drain the preload so the goroutine
		// can complete and not leak past the test.
	}

	// Drain one slot so the blocked send can complete, then wait for the
	// goroutine to finish so we don't leak it across tests.
	<-q
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("blocked count goroutine did not complete after draining a slot")
	}
	// Drain the second message the goroutine enqueued.
	<-q
}

// TestStatsdSenderDeliversToUDP runs the statsd consumer goroutine end-to-end
// against a local UDP listener. It exercises the consumer half of the queue
// (the goroutine started from main as `go StatsdSender(config)`) and the
// wire-format helpers under -race conditions: M producer goroutines push N
// metrics each through the injected queue while the sender reads them.
func TestStatsdSenderDeliversToUDP(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenPacket: %v", err)
	}
	t.Cleanup(func() { pc.Close() })

	host, port, err := net.SplitHostPort(pc.LocalAddr().String())
	if err != nil {
		t.Fatalf("SplitHostPort(%q): %v", pc.LocalAddr().String(), err)
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		t.Fatalf("strconv.Atoi(%q): %v", port, err)
	}
	cfg := &Config{StatsdHost: host, StatsdPort: portNum, StatsdProtocol: "udp"}

	const producers = 4
	const perProducer = 5
	const total = producers * perProducer

	q := make(chan string, total)

	senderDone := make(chan struct{})
	go func() {
		defer close(senderDone)
		statsdSender(cfg, q)
	}()

	var wg sync.WaitGroup
	for i := 0; i < producers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perProducer; j++ {
				increment(q, "connectivity.test", []string{"producer:p", "j:v"})
			}
		}(i)
	}
	wg.Wait()
	// Close the queue so statsdSender's range loop terminates, then wait
	// for the goroutine to exit so the -race detector sees no leaked
	// consumer running into subsequent tests.
	close(q)
	select {
	case <-senderDone:
	case <-time.After(2 * time.Second):
		t.Fatal("statsdSender did not exit after queue close")
	}

	// Read all expected datagrams from the listener.
	if err := pc.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline: %v", err)
	}
	got := 0
	buf := make([]byte, 1024)
	wantPayload := "connectivity.test:1|c|#producer:p,j:v"
	for got < total {
		n, _, err := pc.ReadFrom(buf)
		if err != nil {
			t.Fatalf("ReadFrom after %d/%d datagrams: %v", got, total, err)
		}
		if string(buf[:n]) != wantPayload {
			t.Fatalf("datagram[%d] = %q; want %q", got, string(buf[:n]), wantPayload)
		}
		got++
	}
}

// TestWaitLoopCompletesWhenChecksSucceed exercises the WaitLoop goroutine
// fan-out with a stub check that succeeds on its second invocation per
// destination. The assertion is that wg.Wait() in WaitLoop returns (the test
// itself would time out otherwise) and that every destination's check was
// invoked at least twice — proving the retry loop in WaitFor actually ran.
//
// NumGoroutine bracketing catches the regression of a leaked WaitFor
// goroutine that would silently spin past the test.
func TestWaitLoopCompletesWhenChecksSucceed(t *testing.T) {
	before := runtime.NumGoroutine()

	const destCount = 3

	type destState struct {
		mu       sync.Mutex
		attempts int
	}
	states := make([]*destState, destCount)
	dests := make([]*Destination, destCount)
	checks := make([]func() bool, destCount)
	for i := 0; i < destCount; i++ {
		s := &destState{}
		states[i] = s
		dests[i] = &Destination{Label: "stub", Host: "host", Port: 1}
		checks[i] = func() bool {
			s.mu.Lock()
			defer s.mu.Unlock()
			s.attempts++
			return s.attempts >= 2
		}
	}

	// Drive WaitLoop via the seam: a zero sleep keeps the test fast.
	waitLoop(dests, checks, 0)

	for i, s := range states {
		s.mu.Lock()
		got := s.attempts
		s.mu.Unlock()
		if got < 2 {
			t.Errorf("destination[%d] attempts = %d; want >= 2", i, got)
		}
	}

	// Give scheduler a moment to reap goroutines, then verify no leak.
	// runtime.Gosched is enough — every goroutine returned before waitLoop
	// did.
	runtime.Gosched()
	if after := runtime.NumGoroutine(); after > before {
		t.Errorf("goroutines leaked: before=%d after=%d", before, after)
	}
}

// TestMonitorWithCheckResetsConfidenceOnFailure pins the post-#16 behavior:
// after a success run that saturates confidence at 10 (so the sleep is 10
// minutes), a single failed check snaps the next sleep back to 1 minute. The
// seam exposes one iteration at a time so the test can drive the loop
// synchronously without an injected wall-clock dependency.
//
// Refs #16 — the production logic for the reset already merged; this test
// locks it in so future refactors don't regress it.
func TestMonitorWithCheckResetsConfidenceOnFailure(t *testing.T) {
	dest := &Destination{Label: "stub", Host: "host", Port: 1}

	var sleeps []time.Duration
	sleep := func(d time.Duration) { sleeps = append(sleeps, d) }

	// 12 successes drive confidence past 10 and pin it at the cap, then one
	// failure resets it to 1.
	results := []bool{true, true, true, true, true, true, true, true, true, true, true, true, false}

	confidence := 1
	for _, r := range results {
		r := r
		check := func() bool { return r }
		confidence = dest.monitorWithCheck(confidence, check, sleep)
	}

	if got := len(sleeps); got != len(results) {
		t.Fatalf("sleep call count = %d; want %d", got, len(results))
	}
	// After 10+ successes the saturated sleep should be 10 minutes.
	if got, want := sleeps[10], 10*time.Minute; got != want {
		t.Errorf("sleeps[10] (post-saturation) = %v; want %v", got, want)
	}
	if got, want := sleeps[11], 10*time.Minute; got != want {
		t.Errorf("sleeps[11] (still saturated) = %v; want %v", got, want)
	}
	// The failure at index 12 snaps the next sleep to 1 minute.
	if got, want := sleeps[12], 1*time.Minute; got != want {
		t.Errorf("sleeps[12] (after failure) = %v; want %v", got, want)
	}
	if confidence != 1 {
		t.Errorf("confidence after failure = %d; want 1", confidence)
	}
}
