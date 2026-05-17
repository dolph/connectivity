package main

import "testing"

func TestIcmpProbeSucceeded(t *testing.T) {
	cases := []struct {
		name        string
		packetsRecv int
		probeCount  int
		want        bool
	}{
		{name: "zero_of_three", packetsRecv: 0, probeCount: 3, want: false},
		{name: "one_of_three", packetsRecv: 1, probeCount: 3, want: false},
		{name: "two_of_three", packetsRecv: 2, probeCount: 3, want: true},
		{name: "three_of_three", packetsRecv: 3, probeCount: 3, want: true},
		{name: "two_of_five", packetsRecv: 2, probeCount: 5, want: false},
		{name: "three_of_five", packetsRecv: 3, probeCount: 5, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := icmpProbeSucceeded(tc.packetsRecv, tc.probeCount)
			if got != tc.want {
				t.Fatalf("icmpProbeSucceeded(%d, %d) = %v; want %v", tc.packetsRecv, tc.probeCount, got, tc.want)
			}
		})
	}
}
