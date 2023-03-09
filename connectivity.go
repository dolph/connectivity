package main

import (
	"log"
	"os"
	"sync"
)

/*

This module is responsible for parsing CLI arguments and return codes. It
handles the top-level subcommands and main loops for each (check, wait, and
monitor). All goroutines are managed here.

*/

func main() {
	if len(os.Args) == 1 {
		PrintUsage()
		os.Exit(0)
	}

	command := os.Args[1]

	if command == "validate-config" {
		var configPath string
		var err error
		if len(os.Args) == 3 {
			configPath = os.Args[2]
		} else {
			configPath, err = FindConfig()
			if err != nil {
				log.Fatal(err)
			}
		}

		config := LoadConfig(configPath)
		destinations := ParseDestinations(config.URLs)
		ShowDestinations(destinations)
	} else if command == "check" {
		configPath, _ := FindConfig()
		config := LoadConfig(configPath)
		go StatsdSender(config)
		urls := GetURLs(config)
		destinations := ParseDestinations(urls)
		log.Print("Checking all connectivity...")
		ShowDestinations(destinations)
		if CheckLoop(destinations) {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	} else if command == "wait" || command == "waitfor" {
		configPath, _ := FindConfig()
		config := LoadConfig(configPath)
		go StatsdSender(config)
		urls := GetURLs(config)
		destinations := ParseDestinations(urls)
		ShowDestinations(destinations)
		log.Print("Waiting until all connectivity is validated...")
		WaitLoop(destinations)
	} else if command == "monitor" {
		configPath, _ := FindConfig()
		config := LoadConfig(configPath)
		go StatsdSender(config)
		urls := GetURLs(config)
		destinations := ParseDestinations(urls)
		ShowDestinations(destinations)
		log.Print("Monitoring connectivity...")
		MonitorLoop(destinations)
	} else if command == "version" {
		PrintVersion()
	} else if command == "help" {
		if len(os.Args) == 3 {
			// connectivity help <subcommand>
			if PrintCommandUsage(os.Args[2]) {
				// Command usage for this argument was found
				os.Exit(0)
			} else {
				// Invalid subcommand
				os.Exit(2)
			}
		} else {
			// connectivity help
			PrintUsage()
		}
	} else {
		PrintUsage()

		// Invalid subcommand
		os.Exit(2)
	}
}

func GetURLs(config *Config) []Url {
	if len(os.Args) > 2 {
		// Ignore URLs in the config file and use the ones from the CLI instead
		config.URLs = []Url{}

		for _, url := range os.Args[1:len(os.Args)] {
			config.URLs = append(config.URLs, Url{Url: url})
		}
	}
	return config.URLs
}

func ParseDestinations(urls []Url) []*Destination {
	// Validate all destinations before beginning any monitoring
	errEncountered := false
	var destinations []*Destination
	for idx, url := range urls {
		if idx != 0 {
			dest, err := NewDestination(url)
			if err != nil {
				log.Printf("%s", err)
				errEncountered = true
			} else {
				destinations = append(destinations, dest)
			}
		}
	}
	if errEncountered {
		os.Exit(2)
	}
	return destinations
}

func ShowDestinations(destinations []*Destination) {
	if len(destinations) == 0 {
		log.Print("Failed to parse any destinations")
		os.Exit(1)
	}
	log.Print("Parsed the following destinations:")
	for _, dest := range destinations {
		LogDestination(dest, dest.UrlString())
	}
}

func CheckLoop(destinations []*Destination) bool {
	// Assume all destinations are reachable until proven otherwise
	reachable := true

	// Check destinations sequentially, which is slow, but fixes issue #2
	for _, dest := range destinations {
		if dest.Check() {
			LogDestination(dest, "Connected")
		} else {
			reachable = false
		}
	}

	return reachable
}

func WaitLoop(destinations []*Destination) {
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

func MonitorLoop(destinations []*Destination) {
	for _, dest := range destinations {
		go dest.Monitor()
	}

	// Sleep forever
	select {}
}
