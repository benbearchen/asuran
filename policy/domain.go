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
	delay  *DelayPolicy
	ips    []string
}

type domainPolicyFactory struct {
}

func (*domainPolicyFactory) Keyword() string {
	return domainKeyword
}

func (*domainPolicyFactory) Build(args []string) (Policy, []string, error) {
	var act Policy = nil
	var delay *DelayPolicy = nil
	for len(args) > 0 {
		keyword, rest := args[0], args[1:]
		if !domainSubKeys.isSubKey(keyword) {
			break
		}

		p, rest, err := domainSubKeys.makeSub(keyword, rest)
		if err != nil {
			return nil, rest, err
		} else {
			if d, ok := p.(*DelayPolicy); ok {
				if d.Body() {
					return nil, rest, fmt.Errorf(`domain delay can't be "body": %s`, d.Command())
				}

				delay = d
			} else {
				if act != nil {
					return nil, args, fmt.Errorf(`too many ops: "%s" vs "%s"`, act.Keyword(), p.Keyword())
				}

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
	ips := make([]string, 0)
	if len(args) > 1 {
		for _, address := range strings.Split(args[1], ",") {
			address = strings.TrimSpace(address)
			if address == "" {
				continue
			}

			addr := net.ParseIP(address)
			if addr == nil {
				return nil, nil, fmt.Errorf("invalid ip: %v", args[1])
			} else {
				ips = append(ips, addr.String())
			}
		}
	}

	return newDomainPolicy(target, act, delay, ips), nil, nil
}

func newDomainPolicy(domain string, act Policy, delay *DelayPolicy, ips []string) *DomainPolicy {
	return &DomainPolicy{domain, act, delay, ips}
}

func NewStaticDomainPolicy(domain, ip string) *DomainPolicy {
	return &DomainPolicy{domain, nil, nil, []string{ip}}
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
	if len(d.ips) > 0 {
		cmd = append(cmd, strings.Join(d.ips, ","))
	}

	return strings.Join(cmd, " ")
}

func (d *DomainPolicy) Comment() string {
	suffix := ""
	if d.delay != nil {
		suffix = "(" + d.delay.Comment() + ")"
	}

	act := "正常通行"
	if d.act != nil {
		switch d.act.(type) {
		case *DefaultPolicy:
			act = "正常通行"
		case *BlockPolicy:
			act = "丢弃不返回"
		case *ProxyPolicy:
			act = "代理域名"
		case *NullPolicy:
			act = "查询无结果"
		default:
		}
	}

	return act + suffix
}

func (d *DomainPolicy) Update(p Policy) error {
	switch p := p.(type) {
	case *DomainPolicy:
		d.act = p.act
		d.delay = p.delay
		d.ips = p.ips
	case *DefaultPolicy, *ProxyPolicy, *BlockPolicy, *NullPolicy:
		d.act = p
	case *DelayPolicy:
		d.delay = p
	default:
		return fmt.Errorf("unmatch policy to domain: %s", p.Command())
	}

	return nil
}

func (d *DomainPolicy) Domain() string {
	return d.target
}

func (d *DomainPolicy) Action() Policy {
	return d.act
}

func (d *DomainPolicy) Delete() bool {
	if d.act != nil {
		_, ok := d.act.(*DeletePolicy)
		return ok
	}

	return false
}

func (d *DomainPolicy) SetProxy() {
	d.act = new(ProxyPolicy)
}

func (d *DomainPolicy) Delay() *DelayPolicy {
	return d.delay
}

func (d *DomainPolicy) IP() string {
	if len(d.ips) > 0 {
		return d.ips[0]
	} else {
		return ""
	}
}

func (d *DomainPolicy) IPs() []string {
	return d.ips
}

func (d *DomainPolicy) TargetString() string {
	if len(d.ips) == 0 {
		return ""
	} else {
		return strings.Join(d.ips, ",")
	}
}
