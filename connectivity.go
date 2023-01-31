package main

import (
	"fmt"
	"os"
	"sync"
)

func main() {
	// Configuration
	exitOnSuccess := os.Getenv("C10Y_WAIT")
	fmt.Printf("Wait: %s\n", exitOnSuccess)

	// Basic design:
	// Handoff each URL to a monitoring process that will be in charge of it.
	// Each URL is pre-processed to understand how it can be validated.
	// 1. Can it be parsed into a valid URL?
	// 2. Can it be resolved?
	// 3. Can the host be pinged or each resolved host:port dialed?
	// 4. If it can be reached, handoff to a protocol-specific handler that validates application-level connectivity.
	// 5. Validate the expected outcome. A "permission denied" may be expected from SSH, for example.

	// Validate all destinations before beginning any monitoring
	var destinations []*Destination
	for idx, url := range os.Args {
		if idx != 0 {
			dest, err := NewDestination(url)
			if err != nil {
				os.Exit(2)
			}
			destinations = append(destinations, dest)
		}
	}

	if exitOnSuccess == "1" {
		var wg sync.WaitGroup
		for _, dest := range destinations {
			wg.Add(1)
			go func(dest *Destination) {
				defer wg.Done()
				dest.WaitFor()
			}(dest)
		}

		wg.Wait()
	} else {
		// At this point we know that all destinations are valid, so we can start
		// monitoring each of them.
		for _, dest := range destinations {
			go dest.Monitor()
		}

		// Sleep forever
		select {}
	}
}
