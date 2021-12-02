package net

import (
	. "net"
)

func LocalIPs() []string {
	addrs, err := InterfaceAddrs()
	if err != nil || addrs == nil || len(addrs) == 0 {
		return nil
	}

	ips := make([]string, 0)
	for _, a := range addrs {
		var ip IP
		switch addr := a.(type) {
		case *IPNet:
			ip = addr.IP
		case *IPAddr:
			ip = addr.IP
		default:
			continue
		}

		if !ip.IsGlobalUnicast() {
			continue
		}

		ips = append(ips, ip.String())
	}

	if len(ips) > 0 {
		return ips
	} else {
		return nil
	}
}
