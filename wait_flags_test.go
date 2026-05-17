package main

import (
	"testing"
	"time"
)

func TestParseWaitTimeout(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		want    time.Duration
		wantErr bool
	}{
		{name: "empty", args: nil, want: 0},
		{name: "five_minutes", args: []string{"--timeout", "5m"}, want: 5 * time.Minute},
		{name: "short_flag", args: []string{"-timeout", "30s"}, want: 30 * time.Second},
		{name: "missing_value", args: []string{"--timeout"}, wantErr: true},
		{name: "invalid_duration", args: []string{"--timeout", "nope"}, wantErr: true},
		{name: "non_positive", args: []string{"--timeout", "0s"}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseWaitTimeout(tc.args)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("parseWaitTimeout() = %v; want %v", got, tc.want)
			}
		})
	}
}
