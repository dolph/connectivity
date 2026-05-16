package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// writeConfig writes content to a config file in a hermetic temp dir and
// returns the path. The file (and its parent dir) are torn down by
// t.TempDir's cleanup.
func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "connectivity.yml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q): %v", path, err)
	}
	return path
}

// findURL returns the Url matching label, or nil if not found. It exists so
// table-driven assertions can look up entries without depending on map
// iteration order (the loader is `for k, v := range configMap` over the YAML
// keys).
func findURL(urls []Url, label string) *Url {
	for i := range urls {
		if urls[i].Label == label {
			return &urls[i]
		}
	}
	return nil
}

func TestLoadConfig_EmptyPathReturnsEmptyConfig(t *testing.T) {
	cfg := LoadConfig("")
	if cfg == nil {
		t.Fatalf("LoadConfig(\"\") = nil; want non-nil *Config")
	}
	if len(cfg.URLs) != 0 {
		t.Errorf("LoadConfig(\"\").URLs = %v; want empty", cfg.URLs)
	}
	if cfg.StatsdHost != "" {
		t.Errorf("LoadConfig(\"\").StatsdHost = %q; want empty (defaults are only applied when a path is given)", cfg.StatsdHost)
	}
	if cfg.StatsdPort != 0 {
		t.Errorf("LoadConfig(\"\").StatsdPort = %d; want 0", cfg.StatsdPort)
	}
	if cfg.StatsdProtocol != "" {
		t.Errorf("LoadConfig(\"\").StatsdProtocol = %q; want empty", cfg.StatsdProtocol)
	}
}

func TestLoadConfig_MinimalYAMLAppliesDefaults(t *testing.T) {
	path := writeConfig(t, "example: http://example.com\n")
	cfg := LoadConfig(path)

	if cfg.StatsdHost != "127.0.0.1" {
		t.Errorf("StatsdHost = %q; want %q (default)", cfg.StatsdHost, "127.0.0.1")
	}
	if cfg.StatsdPort != 8125 {
		t.Errorf("StatsdPort = %d; want 8125 (default)", cfg.StatsdPort)
	}
	if cfg.StatsdProtocol != "udp" {
		t.Errorf("StatsdProtocol = %q; want %q (default)", cfg.StatsdProtocol, "udp")
	}
	if len(cfg.URLs) != 1 {
		t.Fatalf("len(URLs) = %d; want 1", len(cfg.URLs))
	}
	got := findURL(cfg.URLs, "example")
	if got == nil {
		t.Fatalf("URL with label %q not found; URLs = %+v", "example", cfg.URLs)
	}
	if got.Url != "http://example.com" {
		t.Errorf("URLs[example].Url = %q; want %q", got.Url, "http://example.com")
	}
}

func TestLoadConfig_MultipleURLs(t *testing.T) {
	yaml := "" +
		"a: http://a.example.com\n" +
		"b: https://b.example.com\n" +
		"c: tcp://c.example.com:1234\n"
	path := writeConfig(t, yaml)
	cfg := LoadConfig(path)

	if len(cfg.URLs) != 3 {
		t.Fatalf("len(URLs) = %d; want 3 — URLs = %+v", len(cfg.URLs), cfg.URLs)
	}
	want := map[string]string{
		"a": "http://a.example.com",
		"b": "https://b.example.com",
		"c": "tcp://c.example.com:1234",
	}
	for label, wantURL := range want {
		got := findURL(cfg.URLs, label)
		if got == nil {
			t.Errorf("URL with label %q not found; URLs = %+v", label, cfg.URLs)
			continue
		}
		if got.Url != wantURL {
			t.Errorf("URLs[%s].Url = %q; want %q", label, got.Url, wantURL)
		}
	}
}

// TestLoadConfig_StatsdKeysBecomeURLs documents the #6 bug: the loader
// unmarshals into map[string]string and treats every YAML key as a URL label,
// so typed config keys like statsd_host / statsd_port / statsd_protocol end up
// in the URLs slice instead of populating the Config struct.
//
// Refs #6 — flip when fixed: once the loader honors the Config struct, these
// keys should populate StatsdHost/StatsdPort/StatsdProtocol and NOT appear in
// URLs. Note that statsd_port's YAML value is an int and would fail to
// unmarshal into map[string]string today, so we use a string-shaped fixture
// the current loader can parse (otherwise the bug manifests as a log.Fatalf
// in LoadConfig rather than a misclassified URL).
func TestLoadConfig_StatsdKeysBecomeURLs(t *testing.T) {
	yaml := "" +
		"statsd_host: \"statsd.example.com\"\n" +
		"statsd_protocol: \"tcp\"\n" +
		"example: \"http://example.com\"\n"
	path := writeConfig(t, yaml)
	cfg := LoadConfig(path)

	// Bug: statsd_host and statsd_protocol are treated as URL labels.
	if got := findURL(cfg.URLs, "statsd_host"); got == nil {
		t.Errorf("expected URL with label %q to be present (current buggy behavior — #6); URLs = %+v", "statsd_host", cfg.URLs)
	} else if got.Url != "statsd.example.com" {
		t.Errorf("URLs[statsd_host].Url = %q; want %q (current buggy behavior — #6)", got.Url, "statsd.example.com")
	}
	if got := findURL(cfg.URLs, "statsd_protocol"); got == nil {
		t.Errorf("expected URL with label %q to be present (current buggy behavior — #6); URLs = %+v", "statsd_protocol", cfg.URLs)
	}

	// Bug: defaults are applied because the typed Config fields were never
	// populated from YAML — the statsd_host value above is silently dropped.
	if cfg.StatsdHost != "127.0.0.1" {
		t.Errorf("StatsdHost = %q; want %q (current buggy behavior — #6: typed YAML keys are dropped, so defaults kick in)", cfg.StatsdHost, "127.0.0.1")
	}
	if cfg.StatsdProtocol != "udp" {
		t.Errorf("StatsdProtocol = %q; want %q (current buggy behavior — #6: typed YAML keys are dropped, so defaults kick in)", cfg.StatsdProtocol, "udp")
	}
}

// TestLoadConfig_MissingFileFatals pins the current log.Fatalf-on-read-error
// behavior. The check uses the helper subprocess pattern: this test re-execs
// the test binary with an environment variable that triggers the helper
// branch to invoke LoadConfig on a non-existent path. We assert the
// subprocess exits non-zero and prints a message naming the file.
//
// The log.Fatalf paths in LoadConfig violate the AGENTS.md guidance
// ("log.Fatalf is acceptable only at process startup; library-level code
// returns errors"). Pinning the current behavior in a test makes future
// refactoring observable.
func TestLoadConfig_MissingFileFatals(t *testing.T) {
	if os.Getenv("CONNECTIVITY_TEST_LOADCONFIG_HELPER") == "1" {
		// Child process: invoke LoadConfig on a path guaranteed not to
		// exist. log.Fatalf will call os.Exit(1).
		LoadConfig(filepath.Join(t.TempDir(), "does-not-exist.yml"))
		// Unreachable when the bug/behavior is intact.
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestLoadConfig_MissingFileFatals$")
	cmd.Env = append(os.Environ(), "CONNECTIVITY_TEST_LOADCONFIG_HELPER=1")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("subprocess exited with status 0; want non-zero. output:\n%s", out)
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("subprocess err = %v (%T); want *exec.ExitError", err, err)
	}
	if exitErr.ExitCode() == 0 {
		t.Errorf("subprocess exit code = 0; want non-zero")
	}
	if !strings.Contains(string(out), "Failed to open config file") {
		t.Errorf("subprocess output = %q; want it to contain %q", string(out), "Failed to open config file")
	}
}

func TestFindConfig_ReturnsErrorWhenNoneExist(t *testing.T) {
	// Run in a temp dir so the current working directory does not contain
	// any of the relative ConfigPaths (connectivity.yml, etc).
	dir := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Fatalf("os.Chdir(%q): %v", prev, err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("os.Chdir(%q): %v", dir, err)
	}

	path, err := FindConfig()
	if err == nil {
		t.Errorf("FindConfig() = (%q, nil); want non-nil error when no config file exists", path)
	}
	if path != "" {
		t.Errorf("FindConfig() path = %q; want empty when error is non-nil", path)
	}
}
