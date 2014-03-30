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

		if ip.IsLoopback() || ip.To4() == nil {
			continue
		}

		ipv4 := ip.String()
		if strings.HasPrefix(ipv4, "169.254.") || strings.HasPrefix(ipv4, "127.0.") || ipv4 == "0.0.0.0" {
			continue
		}

		ips = append(ips, ipv4)
	}

	if len(ips) > 0 {
		return ips
	} else {
		return nil
	}
}
