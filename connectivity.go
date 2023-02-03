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
		var configPath string
		if len(os.Args) == 3 {
			configPath = os.Args[2]
		} else {
			configPath = FindConfig()
		}
		config := LoadConfig(configPath)
		destinations := ParseDestinations(config.URLs)
		ShowDestinations(destinations)
	} else if command == "wait" {
		config := LoadConfig(FindConfig())
		urls := GetURLs(config)
		destinations := ParseDestinations(urls)
		WaitForConnectivity(destinations)
	} else if command == "monitor" {
		config := LoadConfig(FindConfig())
		urls := GetURLs(config)
		destinations := ParseDestinations(urls)
		MonitorConnectivityForever(destinations)
	} else if command == "help" {
		if len(os.Args) == 3 {
			PrintCommandUsage(os.Args[2])
		} else {
			PrintUsage()
		}
	} else {
		PrintUsage()
		os.Exit(1)
	}
}

func PrintUsage() {
	fmt.Println("connectivity is a tool for validating network connectivity requirements.")
	fmt.Println("")
	fmt.Println("Usage: connectivity <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  wait             Wait for all connectivity to be verified at least once")
	fmt.Println("  monitor          Continuously monitor all connectivity forever")
	fmt.Println("  validate-config  Load config without making any network requests")
	fmt.Println("  help             Show this help text")
	fmt.Println("")
	fmt.Println("Use \"connectivity help <command>\" for more information about that command.")
}

func PrintCommandUsage(command string) {
	if command == "wait" {
		fmt.Println("Wait for all specified connectivity to be verified at least once, and exit.")
		fmt.Println("")
		fmt.Println("Usage: connectivity wait [urls]")
		fmt.Println("")
		fmt.Println("This is useful when you need to wait for DNS propogation, a process to start")
		fmt.Println("listening, configuration to be applied, etc, before doing something else.")
	} else if command == "monitor" {
		fmt.Println("Continuously monitor all connectivity forever.")
		fmt.Println("")
		fmt.Println("Usage: connectivity monitor [urls]")
		fmt.Println("")
		fmt.Println("This is useful to run as a daemon for continuously monitoring network")
		fmt.Println("dependencies.")
	} else if command == "validate-config" {
		fmt.Println("Wait for all connectivity to be verified at least once.")
		fmt.Println("")
		fmt.Println("Usage: connectivity validate-config [config-path]")
		fmt.Println("")
		fmt.Println("Any validation errors will produce a non-zero return code (1). Only the config")
		fmt.Println("file at the specified path is validated. If no config file is specified, then")
		fmt.Println("the first config file discovered in order order of precedence is validated:")
		fmt.Println("")
		fmt.Println("- ./connectivity.yml")
		fmt.Println("- ~/.connectivity.yml")
		fmt.Println("- /etc/connectivity.yml")
	} else {
		PrintUsage()
		os.Exit(1)
	}
}

func GetURLs(config *Config) []string {
	if len(os.Args) > 2 {
		return os.Args[1:len(os.Args)]
	}
	return config.URLs
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
