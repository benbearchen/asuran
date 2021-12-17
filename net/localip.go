package net

import (
	. "net"
	"strings"
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

func ShiftRightV4(ips []string) []string {
	if len(ips) == 0 {
		return ips
	}

	a6 := make([]string, 0, len(ips))
	a4 := make([]string, 0, len(ips))
	for _, ip := range ips {
		if strings.IndexByte(ip, ':') >= 0 {
			a6 = append(a6, ip)
		} else {
			a4 = append(a4, ip)
		}
	}

	return append(a6, a4...)
}
