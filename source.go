package main

import "net"

func GetLocalIPs() []string {
	var IPs []string
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, address := range addrs {
			// check the address type and if it is not a loopback the display it
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					IPs = append(IPs, ipnet.IP.String())
				}
			}
		}
	}
	return IPs
}
