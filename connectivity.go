package main

import (
	"log"
	"os"
	"strconv"
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
		if CheckForConnectivityOnce(destinations) {
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
		WaitForConnectivity(destinations)
	} else if command == "monitor" {
		configPath, _ := FindConfig()
		config := LoadConfig(configPath)
		go StatsdSender(config)
		urls := GetURLs(config)
		destinations := ParseDestinations(urls)
		MonitorConnectivityForever(destinations)
	} else if command == "version" {
		PrintVersion()
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

func GetURLs(config *Config) []Url {
	if len(os.Args) > 2 {
		// Ignore URLs in the config file and use the ones from the CLI instead
		config.URLs = []Url{}

		for idx, url := range os.Args[1:len(os.Args)] {
			config.URLs = append(config.URLs, Url{
				Label: strconv.Itoa(idx),
				Url:   url})
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
	for idx, dest := range destinations {
		log.Printf("%d %s %s\n", idx+1, dest, dest.UrlString())
	}
}

func CheckForConnectivityOnce(destinations []*Destination) bool {
	chanOwner := func(dest *Destination) <-chan bool {
		ch := make(chan bool)
		go func(dest *Destination) {
			defer close(ch)
			reachable := dest.Check()
			if reachable {
				log.Printf("%v%v Validated", GetLocalIPs(), dest)
			}
			ch <- reachable
		}(dest)
		return ch
	}

	consumer := func(ch <-chan bool) bool {
		return <-ch
	}

	// Call all channel owners
	checks := []<-chan bool{}
	for _, dest := range destinations {
		checks = append(checks, chanOwner(dest))
	}

	// Collect results from consumers
	results := []bool{}
	for _, check := range checks {
		results = append(results, consumer(check))
	}

	for _, result := range results {
		if !result {
			return false
		}
	}

	return true
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
