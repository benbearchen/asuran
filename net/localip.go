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
		ip := a.String()
		if strings.HasPrefix(ip, "169.254.") || strings.HasPrefix(ip, "127.0.") || ip == "0.0.0.0" {
			continue
		}

		ips = append(ips, ip)
	}

	if len(ips) > 0 {
		return ips
	} else {
		return nil
	}
}
