package main

import (
	"fmt"
	"time"
)

// parseWaitTimeout scans args for --timeout / -timeout and returns the duration.
// A zero duration means wait indefinitely (current default behavior).
func parseWaitTimeout(args []string) (time.Duration, error) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--timeout", "-timeout":
			if i+1 >= len(args) {
				return 0, fmt.Errorf("missing value for %s", args[i])
			}
			d, err := time.ParseDuration(args[i+1])
			if err != nil {
				return 0, fmt.Errorf("invalid timeout %q: %w", args[i+1], err)
			}
			if d <= 0 {
				return 0, fmt.Errorf("timeout must be positive, got %v", d)
			}
			return d, nil
		}
	}
	return 0, nil
}
