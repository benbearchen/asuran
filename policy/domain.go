package policy

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

const domainKeyword = "domain"

var domainSubKeys sub

func init() {
	domainSubKeys.init(
		defaultKeyword,
		blockKeyword,
		proxyKeyword,
		nullKeyword,
		shuffleKeyword,
		circularKeyword,
		nKeyword,
		delayKeyword,
		deleteKeyword,
	)

	regFactory(new(domainPolicyFactory))
}

type DomainPolicy struct {
	target string
	act    Policy
	delay  *DelayPolicy
	opts   map[string]Policy
	ips    []string

	c *domainContext
}

type domainPolicyFactory struct {
}

func (*domainPolicyFactory) Keyword() string {
	return domainKeyword
}

func (*domainPolicyFactory) Build(args []string) (Policy, []string, error) {
	var act Policy = nil
	var delay *DelayPolicy = nil
	opts := make(map[string]Policy)
	for len(args) > 0 {
		keyword, rest := args[0], args[1:]
		if !domainSubKeys.isSubKey(keyword) {
			break
		}

		p, rest, err := domainSubKeys.makeSub(keyword, rest)
		if err != nil {
			return nil, rest, err
		} else {
			switch p := p.(type) {
			case *DelayPolicy:
				if p.Body() {
					return nil, rest, fmt.Errorf(`domain delay can't be "body": %s`, p.Command())
				}

				delay = p
			case *ShufflePolicy, *CircularPolicy, *NPolicy:
				opts[p.Keyword()] = p
			default:
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

	return newDomainPolicy(target, act, delay, opts, ips), nil, nil
}

func newDomainPolicy(domain string, act Policy, delay *DelayPolicy, opts map[string]Policy, ips []string) *DomainPolicy {
	return &DomainPolicy{domain, act, delay, opts, ips, nil}
}

func NewStaticDomainPolicy(domain, ip string) *DomainPolicy {
	return &DomainPolicy{domain, nil, nil, map[string]Policy{}, []string{ip}, nil}
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

	for _, p := range d.opts {
		cmd = append(cmd, p.Command())
	}

	cmd = append(cmd, d.target)
	if len(d.ips) > 0 {
		cmd = append(cmd, strings.Join(d.ips, ","))
	}

	return strings.Join(cmd, " ")
}

func (d *DomainPolicy) Comment() string {
	ex := make([]string, 0)
	if d.delay != nil {
		ex = append(ex, d.delay.Comment())
	}

	for _, p := range d.opts {
		ex = append(ex, p.Comment())
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

	suffix := ""
	if len(ex) != 0 {
		suffix = "(" + strings.Join(ex, ",") + ")"
	}

	return act + suffix
}

func (d *DomainPolicy) Update(p Policy) error {
	switch p := p.(type) {
	case *DomainPolicy:
		d.act = p.act
		d.delay = p.delay
		opts := make(map[string]Policy)
		for k, v := range p.opts {
			opts[k] = v
		}

		d.opts = opts
		d.ips = p.ips
		d.c = nil
	case *DefaultPolicy, *ProxyPolicy, *BlockPolicy, *NullPolicy:
		d.act = p
	case *DelayPolicy:
		d.delay = p
	case *ShufflePolicy, *CircularPolicy, *NPolicy:
		d.opts[p.Keyword()] = p
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

func (d *DomainPolicy) Shuffle() bool {
	_, ok := d.opts[shuffleKeyword]
	return ok
}

func (d *DomainPolicy) Circular() bool {
	_, ok := d.opts[circularKeyword]
	return ok
}

func (d *DomainPolicy) N() (int, bool) {
	p, ok := d.opts[nKeyword]
	if ok {
		return p.(*NPolicy).N(), true
	} else {
		return 0, false
	}
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

type domainContext struct {
	rand *rand.Rand
	in   []string
	out  []string
}

func shuffleStrings(s []string, rnd *rand.Rand) []string {
	r := make([]string, len(s))
	p := rnd.Perm(len(s))
	for i := 0; i < len(r); i++ {
		r[i] = s[p[i]]
	}

	return r
}

func (d *DomainPolicy) NextIPs() []string {
	shuffle := d.Shuffle()
	circular := d.Circular()
	n, hasN := d.N()

	if d.c == nil {
		c := new(domainContext)
		c.rand = rand.New(rand.NewSource(time.Now().UnixNano()))

		d.c = c
	}

	if circular {
		if len(d.c.in) == 0 && len(d.c.out) == 0 {
			d.c.out = d.ips[:]
		}

		size := len(d.ips)
		if hasN && n < size {
			size = n
		}

		if len(d.c.in) < size {
			out := d.c.out
			if shuffle {
				out = shuffleStrings(out, d.c.rand)
			}

			d.c.in = append(d.c.in, out...)
			d.c.out = []string{}
		}

		result := d.c.in[:size]
		d.c.out = append(d.c.out, result...)
		d.c.in = d.c.in[size:]
		return result
	}

	ips := d.ips
	if shuffle {
		ips = shuffleStrings(ips, d.c.rand)
	}

	if hasN && n < len(ips) {
		ips = ips[:n]
	}

	return ips
}

func (d *DomainPolicy) TargetString() string {
	if len(d.ips) == 0 {
		return ""
	} else {
		return strings.Join(d.ips, ",")
	}
}
