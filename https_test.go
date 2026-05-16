package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// newTestDestination returns a *Destination wired to the given URL with a
// label so log calls inside HTTPS don't panic. The Scheme/Host/Port fields
// aren't read by HTTPS itself (it uses dest.URL), but they're populated for
// consistency with what NewDestination would produce.
func newTestDestination(t *testing.T, url string) *Destination {
	t.Helper()
	t.Cleanup(func() { drainQueue(t) })
	drainQueue(t)
	return &Destination{
		Label:    "test",
		URL:      url,
		Protocol: "tcp",
		Scheme:   "http",
		Host:     "example.com",
		Port:     80,
	}
}

func TestHTTPS_Returns200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	dest := newTestDestination(t, srv.URL)
	if !HTTPS(dest) {
		t.Errorf("HTTPS(2xx) = false; want true")
	}
}

// TestHTTPS_Returns500ButReportsSuccess documents that HTTPS does not check
// the response status code, so a 5xx response still returns true. The Go
// stdlib http.Get only returns an error for transport-level failures (DNS,
// dial, TLS, etc.), not for HTTP error statuses.
//
// Refs #7 — flip when fixed: once HTTPS checks status codes, a 5xx response
// should return false. To flip this test then, change `want true` to
// `want false` and update the test name.
func TestHTTPS_Returns500ButReportsSuccess(t *testing.T) {
	cases := []struct {
		name   string
		status int
	}{
		{name: "internal_server_error", status: http.StatusInternalServerError},
		{name: "bad_gateway", status: http.StatusBadGateway},
		{name: "service_unavailable", status: http.StatusServiceUnavailable},
		{name: "not_found", status: http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
			}))
			t.Cleanup(srv.Close)

			dest := newTestDestination(t, srv.URL)
			if !HTTPS(dest) {
				t.Errorf("HTTPS(%d) = false; want true (current buggy behavior — #7: no status-code check)", tc.status)
			}
		})
	}
}

// TestHTTPS_FollowsRedirects documents that HTTPS follows redirects (up to
// Go's default of 10 hops) without restriction. A check tool that silently
// follows a redirect to a different host can mask the very routing /
// availability issues it's meant to detect.
//
// Refs #15 — flip when fixed: once HTTPS disables redirect following (or
// records the redirect chain), this test should assert that the destination
// behind a 3xx is NOT considered reachable, or that the redirect target was
// recorded.
func TestHTTPS_FollowsRedirects(t *testing.T) {
	var finalHits int32
	final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&finalHits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(final.Close)

	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL, http.StatusFound)
	}))
	t.Cleanup(redirector.Close)

	dest := newTestDestination(t, redirector.URL)
	if !HTTPS(dest) {
		t.Errorf("HTTPS(redirect) = false; want true (current buggy behavior — #15: redirects are followed)")
	}
	if got := atomic.LoadInt32(&finalHits); got != 1 {
		t.Errorf("final server hit count = %d; want 1 (current buggy behavior — #15: redirects are followed)", got)
	}
}

// TestHTTPS_DialFailureReturnsFalse pins the only error path HTTPS currently
// surfaces: a transport-level dial failure. This is the one input where the
// current implementation behaves correctly, so the assertion is `want false`
// outright.
func TestHTTPS_DialFailureReturnsFalse(t *testing.T) {
	// 127.0.0.1:1 is a port that's vanishingly unlikely to have a
	// listener; the connection refuses immediately on Linux.
	dest := newTestDestination(t, "http://127.0.0.1:1/")
	if HTTPS(dest) {
		t.Errorf("HTTPS(unreachable) = true; want false")
	}
}
