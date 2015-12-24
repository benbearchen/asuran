package policy

import (
	"net"
)

const hostKeyword = "host"

type HostPolicy struct {
	stringPolicy
}

func init() {
	regFactory(newStringPolicyFactory(hostKeyword, "ip:port", func(val string) (Policy, error) {
		return &HostPolicy{stringPolicy{hostKeyword, val, func(val string) string {
			return "连接 Host " + val
		}}}, nil
	}))
}

func (p *HostPolicy) Host() string {
	return p.Value()
}

func (p *HostPolicy) HTTP() string {
	domain, port, err := net.SplitHostPort(p.Host())
	if err != nil {
		domain = p.Host()
	}

	if len(port) == 0 {
		port = "80"
	}

	return net.JoinHostPort(domain, port)
}
