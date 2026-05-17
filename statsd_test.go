package main

import (
	"strings"
	"testing"
	"time"
)

// drainQueue removes all pending messages from the package-level statsd queue
// so each test starts and ends with an empty channel. Tests that don't enqueue
// anything still call drainQueue via t.Cleanup as a defensive measure against
// state leaking between cases.
func drainQueue(t *testing.T) {
	t.Helper()
	for {
		select {
		case <-queue:
		default:
			return
		}
	}
}

// recvQueue reads one message from the queue with a short timeout. A timeout
// indicates the function under test did not enqueue anything, which is a test
// failure rather than a hang.
func recvQueue(t *testing.T) string {
	t.Helper()
	select {
	case s := <-queue:
		return s
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timed out waiting for message on statsd queue")
		return ""
	}
}

func TestEscapeTag(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "colon", in: "a:b", want: "a-b"},
		{name: "pipe", in: "a|b", want: "a-b"},
		{name: "comma", in: "a,b", want: "a-b"},
		{name: "at", in: "a@b", want: "a-b"},
		{name: "all_specials_combined", in: "a:b|c,d@e", want: "a-b-c-d-e"},
		{name: "no_specials_unchanged", in: "plain.tag_value-1", want: "plain.tag_value-1"},
		{name: "empty_string", in: "", want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EscapeTag(tc.in)
			if got != tc.want {
				t.Errorf("EscapeTag(%q) = %q; want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestEscapeTag_EscapesNewlineAndCR(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "newline", in: "a\nb", want: "a-b"},
		{name: "carriage_return", in: "a\rb", want: "a-b"},
		{name: "crlf", in: "a\r\nb", want: "a--b"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EscapeTag(tc.in)
			if got != tc.want {
				t.Errorf("EscapeTag(%q) = %q; want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestCount_WireFormat(t *testing.T) {
	t.Cleanup(func() { drainQueue(t) })
	drainQueue(t)

	cases := []struct {
		name   string
		metric string
		value  int
		tags   []string
		want   string
	}{
		{
			name:   "single_tag",
			metric: "connectivity.check",
			value:  1,
			tags:   []string{"dest_host:example.com"},
			want:   "connectivity.check:1|c|#dest_host:example.com",
		},
		{
			name:   "multiple_tags_comma_joined",
			metric: "connectivity.check",
			value:  3,
			tags:   []string{"dest_host:example.com", "dest_port:443"},
			want:   "connectivity.check:3|c|#dest_host:example.com,dest_port:443",
		},
		{
			name:   "zero_value",
			metric: "m",
			value:  0,
			tags:   []string{"t:v"},
			want:   "m:0|c|#t:v",
		},
		{
			name:   "negative_value",
			metric: "m",
			value:  -5,
			tags:   []string{"t:v"},
			want:   "m:-5|c|#t:v",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			Count(tc.metric, tc.value, tc.tags)
			got := recvQueue(t)
			if got != tc.want {
				t.Errorf("Count enqueued %q; want %q", got, tc.want)
			}
		})
	}
}

func TestIncrement_EnqueuesCountOfOne(t *testing.T) {
	t.Cleanup(func() { drainQueue(t) })
	drainQueue(t)

	Increment("connectivity.check", []string{"dest_host:example.com"})
	got := recvQueue(t)
	want := "connectivity.check:1|c|#dest_host:example.com"
	if got != want {
		t.Errorf("Increment enqueued %q; want %q", got, want)
	}
}

func TestTimer_WireFormat(t *testing.T) {
	t.Cleanup(func() { drainQueue(t) })
	drainQueue(t)

	cases := []struct {
		name   string
		metric string
		took   time.Duration
		tags   []string
		want   string
	}{
		{
			name:   "whole_millisecond",
			metric: "connectivity.lookup",
			took:   5 * time.Millisecond,
			tags:   []string{"dest_host:example.com"},
			want:   "connectivity.lookup:5|ms|#dest_host:example.com",
		},
		{
			name:   "sub_millisecond_truncates_to_zero",
			metric: "connectivity.lookup",
			took:   500 * time.Microsecond,
			tags:   []string{"t:v"},
			want:   "connectivity.lookup:0|ms|#t:v",
		},
		{
			name:   "second_converts_to_1000ms",
			metric: "m",
			took:   time.Second,
			tags:   []string{"t:v"},
			want:   "m:1000|ms|#t:v",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			Timer(tc.metric, tc.took, tc.tags)
			got := recvQueue(t)
			if got != tc.want {
				t.Errorf("Timer enqueued %q; want %q", got, tc.want)
			}
		})
	}
}

func TestGauge_WireFormat(t *testing.T) {
	t.Cleanup(func() { drainQueue(t) })
	drainQueue(t)

	cases := []struct {
		name   string
		metric string
		value  int
		tags   []string
		want   string
	}{
		{
			name:   "single_tag",
			metric: "connectivity.confidence",
			value:  7,
			tags:   []string{"dest_host:example.com"},
			want:   "connectivity.confidence:7|g|#dest_host:example.com",
		},
		{
			name:   "multiple_tags",
			metric: "g",
			value:  10,
			tags:   []string{"a:1", "b:2", "c:3"},
			want:   "g:10|g|#a:1,b:2,c:3",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			Gauge(tc.metric, tc.value, tc.tags)
			got := recvQueue(t)
			if got != tc.want {
				t.Errorf("Gauge enqueued %q; want %q", got, tc.want)
			}
		})
	}
}

// TestCount_TagSeparatorIsComma verifies the dogstatsd contract that tags are
// comma-separated after `#`. A regression to space-separated tags would still
// look syntactically plausible in logs but silently drop tag parsing at the
// collector.
func TestCount_TagSeparatorIsComma(t *testing.T) {
	t.Cleanup(func() { drainQueue(t) })
	drainQueue(t)

	Count("m", 1, []string{"a:1", "b:2"})
	got := recvQueue(t)
	if !strings.Contains(got, "#a:1,b:2") {
		t.Errorf("Count enqueued %q; want it to contain %q", got, "#a:1,b:2")
	}
}
