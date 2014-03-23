package proxy

import (
	"github.com/benbearchen/asuran/profile"
	"github.com/benbearchen/asuran/util/cmd"
	"github.com/benbearchen/asuran/web/proxy/life"

	"net"
	"strings"
)

func (p *Proxy) Command(commands string, f *profile.Profile, v *life.Life) {
	commandLines := strings.Split(commands, "\n")
	for _, line := range commandLines {
		line = strings.TrimSpace(line)
		if len(line) <= 0 || line[0] == '#' {
			continue
		}

		c, rest := cmd.TakeFirstArg(line)
		switch c {
		case "restart":
			if v != nil {
				v.Restart()
			}
		case "clear":
			f.Clear()
		case "domain":
			f.CommandDomain(rest)
		case "url":
			f.CommandUrl(rest)
		case "host":
			c, rest = cmd.TakeFirstArg(rest)
			if ip, domain, ok := parseIPDomain(c, rest); ok {
				f.CommandDomain(domain + " " + ip)
			}
		default:
			if ip, domain, ok := parseIPDomain(c, rest); ok {
				f.CommandDomain(domain + " " + ip)
			}
		}
	}
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
