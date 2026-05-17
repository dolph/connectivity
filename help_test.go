package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("pipe close: %v", err)
	}
	os.Stdout = old

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("io.ReadAll: %v", err)
	}
	return string(out)
}

func TestPrintCommandUsageValidateConfigListsYamlVariants(t *testing.T) {
	out := captureStdout(t, func() {
		if !PrintCommandUsage("validate-config") {
			t.Fatalf("PrintCommandUsage(validate-config) = false; want true")
		}
	})

	for _, want := range []string{
		"./connectivity.yml",
		"./connectivity.yaml",
		"~/.connectivity.yml",
		"~/.connectivity.yaml",
		"/etc/connectivity.yml",
		"/etc/connectivity.yaml",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("validate-config help missing %q in output:\n%s", want, out)
		}
	}
	if strings.Contains(out, "order order") {
		t.Errorf("validate-config help contains duplicated word: %q", out)
	}
}
