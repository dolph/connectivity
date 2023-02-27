package main

import "log"

func LogDestination(dest *Destination, msg string) {
	log.Printf("%s %s %s", GetLocalIPs(), dest, msg)
}

func LogDestinationError(dest *Destination, msg string, err error) {
	log.Printf("%s %s %s: %s", GetLocalIPs(), dest, msg, err)
}

func LogRoute(route *Route, msg string) {
	log.Printf("%s %s", route, msg)
}

func LogRouteError(route *Route, msg string, err error) {
	log.Printf("%s %s: %s", route, msg, err)
}

func LogRouteDestinationError(route *Route, dest *Destination, msg string, err error) {
	log.Printf("%s %s %s: %s", route, dest, msg, err)
}
