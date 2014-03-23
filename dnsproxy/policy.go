package dnsproxy

import (
	"github.com/benbearchen/asuran/profile"

	_ "fmt"
	"net"
	"strings"
)

type Policy struct {
	op profile.DomainOperator
}

func NewPolicy(op profile.DomainOperator) *Policy {
	p := Policy{}
	p.op = op
	return &p
}

func (p *Policy) Query(clientIP, domain string) (string, []net.IP) {
	pureDomain := domain
	if strings.HasSuffix(domain, ".") {
		pureDomain = domain[0 : len(domain)-1]
	}

	a := p.op.Action(clientIP, pureDomain)
	if a == nil {
		return passDomain(domain, "")
	}

	//fmt.Println(clientIP + " domain " + domain + " " + a.Act.String() + " " + a.TargetString())
	switch a.Act {
	case profile.DomainActNone:
		return passDomain(domain, a.IP)
	case profile.DomainActBlock:
		return domain, nil
	case profile.DomainActProxy:
		return passDomain(domain, a.IP)
	default:
		return passDomain(domain, a.IP)
	}
}

func passDomain(domain, ip string) (string, []net.IP) {
	if len(ip) > 0 {
		return domain, []net.IP{net.ParseIP(ip)}
	} else {
		ips, err := querySystemDns(domain)
		if err != nil {
			return domain, nil
		}

		return domain, ips
	}
}
