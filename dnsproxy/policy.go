package dnsproxy

import (
	"github.com/benbearchen/asuran/policy"

	_ "fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

type DomainOperator interface {
	Action(ip, domain string) *policy.DomainPolicy
}

type Policy struct {
	op DomainOperator
	r  *rand.Rand
}

func NewPolicy(op DomainOperator) *Policy {
	p := Policy{}
	p.op = op
	p.r = rand.New(rand.NewSource(time.Now().UnixNano()))
	return &p
}

func (p *Policy) Query(clientIP, domain string) (string, []net.IP) {
	pureDomain := domain
	if strings.HasSuffix(domain, ".") {
		pureDomain = domain[0 : len(domain)-1]
	}

	a := p.op.Action(clientIP, pureDomain)
	if a == nil {
		return passDomain(domain, []string{})
	}

	if d := a.Delay(); d != nil {
		duration := d.RandDuration(p.r)
		time.Sleep(duration)
	}

	if a.Action() == nil {
		return passDomain(domain, a.NextIPs())
	}

	//fmt.Println(clientIP + " domain " + domain + " " + a.Act.String() + " " + a.TargetString())
	switch a.Action().(type) {
	case *policy.BlockPolicy:
		return domain, nil
	case *policy.ProxyPolicy:
		return passDomain(domain, a.NextIPs())
	case *policy.NullPolicy:
		return domain, []net.IP{}
	default:
		return passDomain(domain, a.NextIPs())
	}
}

func passDomain(domain string, ips []string) (string, []net.IP) {
	if len(ips) > 0 {
		netIPs := make([]net.IP, 0, len(ips))
		for _, ip := range ips {
			netIPs = append(netIPs, net.ParseIP(ip))
		}

		return domain, netIPs
	} else {
		ips, err := querySystemDns(domain)
		if err != nil {
			return domain, nil
		}

		return domain, ips
	}
}
