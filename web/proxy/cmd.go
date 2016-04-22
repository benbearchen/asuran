package proxy

import (
	"github.com/benbearchen/asuran/policy"
	"github.com/benbearchen/asuran/profile"
	"github.com/benbearchen/asuran/util/cmd"
	"github.com/benbearchen/asuran/web/proxy/life"

	"fmt"
	"net"
	"strings"
)

func (p *Proxy) Command(commands string, f *profile.Profile, v *life.Life) []string {
	ps, errors := p.ParseCommand(commands)
	for _, p := range ps {
		switch p := p.(type) {
		case *policy.RestartPolicy:
			if v != nil {
				v.Restart()
			}
		case *policy.ClearPolicy:
			f.Clear()
		case *policy.DomainPolicy:
			f.SetDomainPolicy(p)
		case *policy.UrlPolicy:
			f.SetUrlPolicy(p)
		default:
		}
	}

	return errors
}

func (*Proxy) ParseCommand(commands string) ([]policy.Policy, []string) {
	errors := make([]string, 0)
	ps := make([]policy.Policy, 0)

	commandLines := strings.Split(commands, "\n")
	for _, line := range commandLines {
		line = strings.TrimSpace(line)
		if len(line) <= 0 || line[0] == '#' {
			continue
		}

		p, err := policy.Factory(line)
		if err != nil {
			c, rest := cmd.TakeFirstArg(line)
			if ip, domain, ok := parseIPDomain(c, rest); ok {
				p = policy.NewStaticDomainPolicy(domain, ip)
			} else {
				errors = append(errors, fmt.Sprintf("%s\n##\t%v\n", line, err))
				continue
			}
		}

		ps = append(ps, p)
	}

	return ps, errors
}

func parseIPDomain(c, rest string) (string, string, bool) {
	ip := net.ParseIP(c)
	if ip != nil {
		d, r := cmd.TakeFirstArg(rest)
		if len(d) > 0 && len(r) == 0 {
			return c, d, true
		}
	}

	return "", "", false
}
