package main

import (
	"fmt"
	"log"
	"os"
	"sync"
)

func main() {
	if len(os.Args) == 1 {
		PrintUsage()
		os.Exit(0)
	}

	command := os.Args[1]

	if command == "validate-config" {
		config := LoadConfig()
		destinations := ParseDestinations(config.URLs)
		ShowDestinations(destinations)
	} else if command == "wait" {
		config := LoadConfig()
		destinations := ParseDestinations(config.URLs)
		WaitForConnectivity(destinations)
	} else if command == "monitor" {
		config := LoadConfig()
		destinations := ParseDestinations(config.URLs)
		MonitorConnectivityForever(destinations)
	} else if command == "help" || command == "--help" || command == "-h" {
		PrintUsage()
	} else {
		PrintUsage()
		os.Exit(1)
	}
}

func PrintUsage() {
	fmt.Println("connectivity is a tool for verifying and debugging network connectivity issues.")
	fmt.Println("")
	fmt.Println("Usage: connectivity <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  wait				Wait for all connectivity to be verified at least once")
	fmt.Println("  monitor			Continuously monitor all connectivity forever")
	fmt.Println("  validate-config  Load config without making any network requests")
	fmt.Println("  help				Show this help text")
}

func ParseDestinations(urls []string) []*Destination {
	// Validate all destinations before beginning any monitoring
	var destinations []*Destination
	for idx, url := range urls {
		if idx != 0 {
			dest, err := NewDestination(url)
			if err != nil {
				os.Exit(2)
			}
			destinations = append(destinations, dest)
		}
	}
	return destinations
}

func ShowDestinations(destinations []*Destination) {
	if len(destinations) == 0 {
		log.Print("Failed to parse any destinations.")
		os.Exit(1)
	}
	log.Print("Parsed the following destinations:")
	for idx, dest := range destinations {
		log.Printf("%d. %s\n", idx+1, dest)
	}
}

func WaitForConnectivity(destinations []*Destination) {
	var wg sync.WaitGroup
	for _, dest := range destinations {
		wg.Add(1)
		go func(dest *Destination) {
			defer wg.Done()
			dest.WaitFor()
		}(dest)
	}

	wg.Wait()
}

func MonitorConnectivityForever(destinations []*Destination) {
	for _, dest := range destinations {
		go dest.Monitor()
	}

	// Sleep forever
	select {}
}
