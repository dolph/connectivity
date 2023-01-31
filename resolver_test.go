package main

import (
	"testing"
)

func assertLookupEquals(t *testing.T, got []string, want []string) {
	if len(got) != len(want) {
		t.Errorf("len(Lookup(dest)) != %d; want %d", len(got), len(want))
	}
	for _, wantResult := range want {
		assertIpInResult(t, got, wantResult)
	}
}

func assertIpInResult(t *testing.T, got []string, want string) bool {
	for _, gotResult := range got {
		if gotResult == want {
			return true
		}
	}
	t.Errorf("Lookup(dest) = %s; missing %s", got, want)
	return false
}

func assertLenOfResultsInRange(t *testing.T, got []string, wantMin int, wantMax int) {
	if wantMin > wantMax {
		t.Errorf("wantMin (%d) must be less than or equal to wantMax (%d)", wantMin, wantMax)
	} else if len(got) < wantMin {
		t.Errorf("len(Lookup(dest)) = %s; want >= %d", got, wantMin)
	} else if len(got) > wantMax {
		t.Errorf("len(Lookup(dest)) = %s; want <= %d", got, wantMax)
	}
}

func TestLookupLoopbackIp(t *testing.T) {
	dest, err := NewDestination("http://127.0.0.1")
	assertNoError(t, dest, err)
	got := Lookup(dest)
	assertLookupEquals(t, got, []string{"127.0.0.1"})
}

func TestLookupPublicIp(t *testing.T) {
	dest, err := NewDestination("http://8.8.8.8")
	assertNoError(t, dest, err)
	got := Lookup(dest)
	assertLookupEquals(t, got, []string{"8.8.8.8"})
}

func TestLookupGoogle(t *testing.T) {
	dest, err := NewDestination("https://google.com")
	assertNoError(t, dest, err)
	got := Lookup(dest)
	assertLenOfResultsInRange(t, got, 2, 20)
}

func TestLookupInvalidHostname(t *testing.T) {
	dest, err := NewDestination("https://a.b.c")
	assertNoError(t, dest, err)
	got := Lookup(dest)
	assertLenOfResultsInRange(t, got, 0, 0)
}
