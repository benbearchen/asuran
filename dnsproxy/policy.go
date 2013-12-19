package dnsproxy

import (
	"github.com/benbearchen/asuran/profile"

	"fmt"
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
	fmt.Println(clientIP + " domain " + domain + " " + a.Act.String() + " " + a.TargetString())
	switch a.Act {
	case profile.DomainActNone:
		return passDomain(domain)
	case profile.DomainActBlock:
		return domain, nil
	case profile.DomainActRedirect:
		return domain, []net.IP{net.ParseIP(a.IP)}
	default:
		return passDomain(domain)
	}
}

func passDomain(domain string) (string, []net.IP) {
	ips, err := querySystemDns(domain)
	if err != nil {
		return domain, nil
	} else {
		return domain, ips
	}
}
