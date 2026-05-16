package main

import (
	"fmt"
	"strings"
	"testing"
)

// testingT is the subset of *testing.T used by the assertion helpers in this
// file. Defining it lets meta-tests substitute a capturing fake so we can
// assert on the failure messages the helpers emit (see TestAssert*).
type testingT interface {
	Errorf(format string, args ...interface{})
	Helper()
}

// fakeT captures a single failure message so meta-tests can assert that
// helpers print a useful diagnostic when they fire. Only the subset of
// *testing.T surface used by the assertion helpers is implemented.
type fakeT struct {
	failed bool
	msg    string
}

func (f *fakeT) Errorf(format string, args ...interface{}) {
	f.failed = true
	f.msg = fmt.Sprintf(format, args...)
}

func (f *fakeT) Helper() {}

// assertPortEquals takes the int field under test directly so the failure
// message can format it with %d and refer to "Port" by name.
func assertPortEquals(t testingT, got int, want int) {
	t.Helper()
	if got != want {
		t.Errorf("Port = %d; want %d", got, want)
	}
}

// assertSchemeEquals takes the string field under test directly so the failure
// message can print the actual scheme that mismatched.
func assertSchemeEquals(t testingT, got string, want string) {
	t.Helper()
	if got != want {
		t.Errorf("Scheme = %q; want %q", got, want)
	}
}

// assertHostEquals takes the string field under test directly so the failure
// message can print the actual host that mismatched.
func assertHostEquals(t testingT, got string, want string) {
	t.Helper()
	if got != want {
		t.Errorf("Host = %q; want %q", got, want)
	}
}

// assertNoError fails the test if err is non-nil. The name parameter labels
// the call site so failures in table-driven tests point at the right row.
func assertNoError(t testingT, name string, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", name, err)
	}
}

// assertError fails the test if err is nil. Prefer assertErrorContains when
// the caller knows which error to expect; this helper exists for the cases
// where the test only cares that an error was returned at all.
func assertError(t testingT, name string, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected an error; got nil", name)
	}
}

// assertErrorContains fails the test if err is nil or if err.Error() does not
// contain substr. Use this instead of assertError to lock in which error a
// failing path is expected to return (avoiding "checklist theater" — see #46).
func assertErrorContains(t testingT, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Errorf("expected an error containing %q; got nil", substr)
		return
	}
	if !strings.Contains(err.Error(), substr) {
		t.Errorf("error = %q; want substring %q", err.Error(), substr)
	}
}

// errString is a tiny error type used to build errors with known messages for
// the helper meta-tests, without depending on production code paths.
type errString string

func (e errString) Error() string { return string(e) }

func TestAssertPortEquals_FailureMessageNamesPort(t *testing.T) {
	ft := &fakeT{}
	assertPortEquals(ft, 80, 443)
	if !ft.failed {
		t.Fatalf("assertPortEquals(80, 443) did not fail; want failure")
	}
	if !strings.Contains(ft.msg, "Port") {
		t.Errorf("message = %q; want it to mention %q", ft.msg, "Port")
	}
	if !strings.Contains(ft.msg, "80") || !strings.Contains(ft.msg, "443") {
		t.Errorf("message = %q; want both got (80) and want (443) to appear", ft.msg)
	}
}

func TestAssertPortEquals_NoFailureWhenEqual(t *testing.T) {
	ft := &fakeT{}
	assertPortEquals(ft, 443, 443)
	if ft.failed {
		t.Errorf("assertPortEquals(443, 443) reported failure %q; want no failure", ft.msg)
	}
}

func TestAssertSchemeEquals_FailureMessageNamesScheme(t *testing.T) {
	ft := &fakeT{}
	assertSchemeEquals(ft, "http", "https")
	if !ft.failed {
		t.Fatalf("assertSchemeEquals(http, https) did not fail; want failure")
	}
	if !strings.Contains(ft.msg, "Scheme") {
		t.Errorf("message = %q; want it to mention %q", ft.msg, "Scheme")
	}
	if !strings.Contains(ft.msg, "http") || !strings.Contains(ft.msg, "https") {
		t.Errorf("message = %q; want both got (http) and want (https) to appear", ft.msg)
	}
}

func TestAssertHostEquals_FailureMessageNamesHost(t *testing.T) {
	ft := &fakeT{}
	assertHostEquals(ft, "example.com", "example.org")
	if !ft.failed {
		t.Fatalf("assertHostEquals failed-case did not fail; want failure")
	}
	if !strings.Contains(ft.msg, "Host") {
		t.Errorf("message = %q; want it to mention %q", ft.msg, "Host")
	}
	if !strings.Contains(ft.msg, "example.com") || !strings.Contains(ft.msg, "example.org") {
		t.Errorf("message = %q; want both got and want hosts to appear", ft.msg)
	}
}

func TestAssertErrorContains_PassesWhenSubstringMatches(t *testing.T) {
	ft := &fakeT{}
	assertErrorContains(ft, errString("port required"), "port required")
	if ft.failed {
		t.Errorf("assertErrorContains with matching substring failed; got message %q", ft.msg)
	}
}

func TestAssertErrorContains_FailsWhenSubstringMissing(t *testing.T) {
	ft := &fakeT{}
	assertErrorContains(ft, errString("out of memory"), "port required")
	if !ft.failed {
		t.Fatalf("assertErrorContains with non-matching substring did not fail; want failure")
	}
	if !strings.Contains(ft.msg, "port required") {
		t.Errorf("message = %q; want it to mention the expected substring %q", ft.msg, "port required")
	}
	if !strings.Contains(ft.msg, "out of memory") {
		t.Errorf("message = %q; want it to include the actual error %q", ft.msg, "out of memory")
	}
}

func TestAssertErrorContains_FailsOnNilError(t *testing.T) {
	ft := &fakeT{}
	assertErrorContains(ft, nil, "port required")
	if !ft.failed {
		t.Errorf("assertErrorContains with nil error did not fail; want failure")
	}
}

func TestAssertNoError_FailureMessageIncludesName(t *testing.T) {
	ft := &fakeT{}
	assertNoError(ft, "tcp://host:123", errString("boom"))
	if !ft.failed {
		t.Fatalf("assertNoError with non-nil error did not fail; want failure")
	}
	if !strings.Contains(ft.msg, "tcp://host:123") {
		t.Errorf("message = %q; want it to mention the call-site name", ft.msg)
	}
	if !strings.Contains(ft.msg, "boom") {
		t.Errorf("message = %q; want it to include the underlying error", ft.msg)
	}
}

func TestMinimalHttpUrl(t *testing.T) {
	got, err := NewDestination(Url{Label: "host", Url: "http://host"})
	assertNoError(t, "http://host", err)
	assertSchemeEquals(t, got.Scheme, "http")
	assertHostEquals(t, got.Host, "host")
	assertPortEquals(t, got.Port, 80)
}

func TestMinimalHttpsUrl(t *testing.T) {
	got, err := NewDestination(Url{Label: "host", Url: "https://host"})
	assertNoError(t, "https://host", err)
	assertSchemeEquals(t, got.Scheme, "https")
	assertHostEquals(t, got.Host, "host")
	assertPortEquals(t, got.Port, 443)
}

func TestMinimalMysqlUrl(t *testing.T) {
	got, err := NewDestination(Url{Label: "mysql_host", Url: "mysql://host"})
	assertNoError(t, "mysql://host", err)
	assertSchemeEquals(t, got.Scheme, "mysql")
	assertHostEquals(t, got.Host, "host")
	assertPortEquals(t, got.Port, 3306)
}

func TestMinimalPostgresUrl(t *testing.T) {
	got, err := NewDestination(Url{Label: "postgres_host", Url: "postgres://host"})
	assertNoError(t, "postgres://host", err)
	assertSchemeEquals(t, got.Scheme, "postgres")
	assertHostEquals(t, got.Host, "host")
	assertPortEquals(t, got.Port, 5432)
}

func TestMinimalNatsUrl(t *testing.T) {
	got, err := NewDestination(Url{Label: "nats_host", Url: "nats://host"})
	assertNoError(t, "nats://host", err)
	assertSchemeEquals(t, got.Scheme, "nats")
	assertHostEquals(t, got.Host, "host")
	assertPortEquals(t, got.Port, 4222)
}

func TestSchemeNormalization(t *testing.T) {
	got, err := NewDestination(Url{Label: "schemy_host", Url: "HTtP://host"})
	assertNoError(t, "HTtP://host", err)
	assertSchemeEquals(t, got.Scheme, "http")
	assertHostEquals(t, got.Host, "host")
	assertPortEquals(t, got.Port, 80)
}

func TestTcpUrlWithoutPort(t *testing.T) {
	_, err := NewDestination(Url{Label: "tcp_host", Url: "tcp://host"})
	assertErrorContains(t, err, "Unsupported scheme")
}

func TestTcpUrlWithPort(t *testing.T) {
	got, err := NewDestination(Url{Label: "tcp_host", Url: "tcp://host:123"})
	assertNoError(t, "tcp://host:123", err)
	assertSchemeEquals(t, got.Scheme, "tcp")
	assertHostEquals(t, got.Host, "host")
	assertPortEquals(t, got.Port, 123)
}

func TestUdpUrlWithoutPort(t *testing.T) {
	_, err := NewDestination(Url{Label: "udp_host", Url: "udp://host"})
	assertErrorContains(t, err, "Unsupported scheme")
}

func TestUdpUrlWithPort(t *testing.T) {
	got, err := NewDestination(Url{Label: "udp_host", Url: "udp://host:123"})
	assertNoError(t, "udp://host:123", err)
	assertSchemeEquals(t, got.Scheme, "udp")
	assertHostEquals(t, got.Host, "host")
	assertPortEquals(t, got.Port, 123)
}
