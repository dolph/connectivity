package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoctorHTTPS_StatusAndBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok-body"))
	}))
	t.Cleanup(srv.Close)

	dest := newTestDestination(t, srv.URL)
	dest.Scheme = "http"

	if !doctorHTTPS(dest) {
		t.Fatal("doctorHTTPS() = false; want true")
	}
}

func TestDoctorHTTPS_Non2xxFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	dest := newTestDestination(t, srv.URL)
	dest.Scheme = "http"

	if doctorHTTPS(dest) {
		t.Fatal("doctorHTTPS() = true; want false for 503")
	}
}
