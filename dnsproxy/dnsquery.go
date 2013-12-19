package dnsproxy

import (
	"net"
)

func querySystemDns(domain string) ([]net.IP, error) {
	return net.LookupIP(domain)
}
