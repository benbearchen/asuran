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
		return passDomain(domain, "")
	}

	if d := a.Delay(); d != nil {
		duration := d.RandDuration(p.r)
		time.Sleep(duration)
	}

	if a.Action() == nil {
		return passDomain(domain, a.IP())
	}

	//fmt.Println(clientIP + " domain " + domain + " " + a.Act.String() + " " + a.TargetString())
	switch a.Action().(type) {
	case *policy.BlockPolicy:
		return domain, nil
	case *policy.ProxyPolicy:
		return passDomain(domain, a.IP())
	case *policy.NullPolicy:
		return domain, []net.IP{}
	default:
		return passDomain(domain, a.IP())
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
