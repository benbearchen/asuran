package policy

import (
	"fmt"
	"net"
	"strings"
)

const domainKeyword = "domain"

var domainSubKeys sub

func init() {
	domainSubKeys.init(
		defaultKeyword,
		blockKeyword,
		proxyKeyword,
		nullKeyword,
		delayKeyword,
		deleteKeyword,
	)

	regFactory(new(domainPolicyFactory))
}

type DomainPolicy struct {
	target string
	act    Policy
	delay  Policy
	ip     string
}

type domainPolicyFactory struct {
}

func (*domainPolicyFactory) Keyword() string {
	return domainKeyword
}

func (*domainPolicyFactory) Build(args []string) (Policy, []string, error) {
	var act, delay Policy
	for len(args) > 0 {
		keyword, rest := args[0], args[1:]
		if !domainSubKeys.isSubKey(keyword) {
			break
		}

		p, rest, err := domainSubKeys.makeSub(keyword, rest)
		if err != nil {
			return nil, rest, err
		} else {
			if _, ok := p.(*DelayPolicy); ok {
				delay = p
			} else {
				act = p
			}
		}

		args = rest
	}

	if len(args) == 0 {
		return nil, args, fmt.Errorf("missing target domain or pattern")
	} else if len(args) > 2 {
		return nil, args, fmt.Errorf("too many args: %v", args)
	}

	target := args[0]
	var ip = ""
	if len(args) > 1 {
		addr := net.ParseIP(args[1])
		if addr == nil {
			return nil, nil, fmt.Errorf("invalid ip: %v", args[1])
		} else {
			ip = addr.String()
		}
	}

	return newDomainPolicy(target, act, delay, ip), nil, nil
}

func newDomainPolicy(domain string, act, delay Policy, ip string) *DomainPolicy {
	return &DomainPolicy{domain, act, delay, ip}
}

func (d *DomainPolicy) Keyword() string {
	return domainKeyword
}

func (d *DomainPolicy) Command() string {
	cmd := make([]string, 0, 5)
	cmd = append(cmd, d.Keyword())
	if d.act != nil {
		if _, ok := d.act.(*DefaultPolicy); !ok {
			cmd = append(cmd, d.act.Keyword())
		}
	}

	if d.delay != nil {
		cmd = append(cmd, d.delay.Command())
	}

	cmd = append(cmd, d.target)
	if len(d.ip) > 0 {
		cmd = append(cmd, d.ip)
	}

	return strings.Join(cmd, " ")
}

func (d *DomainPolicy) Comment() string {
	if d.act == nil {
		return ""
	}

	switch d.act.(type) {
	case *DefaultPolicy:
		return "正常通行"
	case *BlockPolicy:
		return "丢弃不返回"
	case *ProxyPolicy:
		return "代理域名"
	case *NullPolicy:
		return "查询无结果"
	default:
		return ""
	}
}

func (d *DomainPolicy) Update(p Policy) error {
	if d.Keyword() != p.Keyword() {
		return fmt.Errorf("unmatch keywrod: %s vs %s", d.Keyword(), p.Keyword())
	}

	switch p := p.(type) {
	case *DomainPolicy:
		d.target = p.target
		d.act = p.act
		d.delay = p.delay
		d.ip = p.ip
	default:
		return fmt.Errorf("unmatch policy")
	}

	return nil
}

func (d *DomainPolicy) Domain() string {
	return d.target
}

func (d *DomainPolicy) Action() Policy {
	return d.act
}

func (d *DomainPolicy) Delay() Policy {
	return d.delay
}

func (d *DomainPolicy) IP() string {
	return d.ip
}
