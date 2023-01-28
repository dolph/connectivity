package main

import (
	"fmt"
	"log"
	"net"
)

func Dial(dest *Destination, ip string) bool {
	hostPort := fmt.Sprintf("%s:%d", ip, dest.Port)
	// log.Printf("%s Dialing %s://%s", GetLocalIPs(), dest.Protocol, hostPort)

	conn, err := net.Dial(dest.Protocol, hostPort)
	if err != nil {
		log.Printf("%s Failed to dial %s://%s (%v)", GetLocalIPs(), dest.Protocol, hostPort, err)
		return false
	}
	defer conn.Close()
	return true
}
