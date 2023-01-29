package main

import (
	"os"
)

func main() {
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
			destinations = append(destinations, NewDestination(url))
		}
	}

	// At this point we know that all destinations are valid, so we can start
	// monitoring each of them.
	for _, dest := range destinations {
		go dest.Monitor()
	}

	// Sleep forever
	select {}
}
