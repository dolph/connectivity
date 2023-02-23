package main

import (
	"fmt"
	"os"
)

var GitTag string
var GitCommit string
var GoVersion string
var BuildTimestamp string
var BuildOS string
var BuildArch string
var BuildTainted string

func PrintVersion() {
	if GitTag != "" {
		if BuildTainted == "true" {
			fmt.Printf("Version: %s (tainted)\n", GitTag)
		} else {
			fmt.Printf("Version: %s\n", GitTag)
		}
	}
	fmt.Printf("Commit SHA: %s\n", GitCommit)
	fmt.Printf("Go Version: %s\n", GoVersion)
	fmt.Printf("Built: %s\n", BuildTimestamp)
	fmt.Printf("OS/Arch: %s/%s\n", BuildOS, BuildArch)
}

func PrintUsage() {
	fmt.Println("connectivity is a tool for validating network connectivity requirements.")
	fmt.Println("")
	fmt.Println("Usage: connectivity <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  check            Check all connectivity once and exit")
	fmt.Println("  wait             Wait for all connectivity to be validated successfully")
	fmt.Println("  monitor          Continuously monitor all connectivity forever")
	fmt.Println("  validate-config  Load config without making any network requests")
	fmt.Println("  version          Show version information")
	fmt.Println("  help             Show this help text")
	fmt.Println("")
	fmt.Println("Use \"connectivity help <command>\" for more information about that command.")
}
func PrintCommandUsage(command string) {
	if command == "check" {
		fmt.Println("Validate specified connectivity once and exit.")
		fmt.Println("")
		fmt.Println("Usage: connectivity check [urls]")
		fmt.Println("")
		fmt.Println("This is useful when you want to externally orchestrate other processes by")
		fmt.Println("quickly validating connectivity.")
	} else if command == "wait" || command == "waitfor" {
		fmt.Println("Wait for all specified connectivity to be validated successfully at least once.")
		fmt.Println("")
		fmt.Println("Usage: connectivity wait [urls]")
		fmt.Println("")
		fmt.Println("This is useful when you need to wait for DNS propogation, a process to start")
		fmt.Println("listening, configuration to be applied, etc, before doing something else. The")
		fmt.Println("results of each check are emitted via statsd.")
	} else if command == "monitor" {
		fmt.Println("Continuously monitor all connectivity forever.")
		fmt.Println("")
		fmt.Println("Usage: connectivity monitor [urls]")
		fmt.Println("")
		fmt.Println("This is useful to run as a daemon for continuously monitoring network")
		fmt.Println("dependencies. The results of each check are emitted via statsd.")
	} else if command == "version" {
		fmt.Println("Show version information about this build")
		fmt.Println("")
		fmt.Println("Usage: connectivity version")
	} else if command == "validate-config" {
		fmt.Println("Load config without making any network requests.")
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
