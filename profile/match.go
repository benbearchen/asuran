package profile

import (
	"net"
	"regexp"
	"strings"
)

func DomainPatternUsage() string {
	return `domain pattern support '*'

([^.*]+).     match \1\.
*([^.*]+)*.   match [^.]*\1[^.]*\.
*.            match ([^.]+\.)+ or ([^.]+\.)*
`
}

type DomainPattern struct {
	pattern string
	regex   *regexp.Regexp
}

func NewDomainPattern(pattern string) *DomainPattern {
	p := new(DomainPattern)
	p.pattern = pattern
	if strings.Index(pattern, "*") >= 0 {
		regex := domainPattern2Regex(pattern)
		r, err := regexp.Compile(regex)
		if err == nil {
			p.regex = r
		}
	}

	return p
}

func domainPattern2Regex(pattern string) string {
	pattern = strings.TrimSpace(pattern)
	if strings.HasSuffix(pattern, ".") {
		pattern = pattern[0 : len(pattern)-1]
	}

	r := ""
	dots := strings.Split(pattern, ".")
	for i, p := range dots {
		p = strings.TrimSpace(p)
		if len(p) == 0 {
			// ".." ? just pass
		} else if p == "*" {
			if i+1 == len(dots) {
				r += "[^.]+"
			} else {
				rep := "+"
				if len(r) == 0 {
					rep = "*"
				}

				r += "([^.]+\\.)" + rep
			}
		} else if strings.Index(p, "*") >= 0 {
			suffix := "\\."
			if i+1 == len(dots) {
				suffix = ""
			}

			r += strings.Replace(p, "*", "[^.]*", -1) + suffix
		} else {
			suffix := "\\."
			if i+1 == len(dots) {
				suffix = ""
			}

			r += p + suffix
		}
	}

	r = "^" + r + "\\.?$"
	return r
}

func (d *DomainPattern) Match(domain string) bool {
	if d.regex != nil {
		return d.regex.MatchString(domain)
	} else {
		return d.pattern == domain
	}
}

func (d *DomainPattern) MatchScore(domain string) uint8 {
	if d.regex != nil {
		if d.regex.MatchString(domain) {
			return 1
		}
	} else if d.pattern == domain {
		return 2
	}

	return 0
}

func PathPatternUsage() string {
	return `path pattern support '*'

/([^/*]+)     match /\1
/*([^/*]+)*   match /[^/]*\1[^/]*
/*            match (/[^/]+)+ or /([^/]+/)*[^/]*
`
}

type PathPattern struct {
	pattern string
	regex   *regexp.Regexp
}

func NewPathPattern(pattern string) *PathPattern {
	p := new(PathPattern)
	p.pattern = pattern
	if strings.Index(pattern, "*") >= 0 {
		regex := pathPattern2Regex(pattern)
		r, err := regexp.Compile(regex)
		if err == nil {
			p.regex = r
		}
	}

	return p
}

func uniqueTrim(s string, u rune) string {
	r := make([]rune, 0, len(s))
	var last rune = 0
	for _, c := range s {
		if c != u || c != last {
			r = append(r, c)
		}

		last = c
	}

	return string(r)
}

func pathPattern2Regex(pattern string) string {
	pattern = uniqueTrim(strings.TrimSpace(pattern), '/')

	r := ""
	nodes := strings.Split(pattern, "/")
	for i, p := range nodes {
		p = strings.TrimSpace(p)
		if len(p) == 0 {
			// pass
			if i+1 == len(nodes) {
				r += "/"
			}
		} else if p == "*" {
			if i+1 == len(nodes) {
				// end with `/*'
				r += "/([^/]+/)*[^/]*"
			} else {
				r += "(/[^/]+)+"
			}
		} else if strings.Index(p, "*") >= 0 {
			r += "/" + strings.Replace(p, "*", "[^/]*", -1)
		} else {
			r += "/" + p
		}
	}

	r = "^" + r + "$"
	return r
}

func (p *PathPattern) Match(path string) bool {
	if p.regex != nil {
		return p.regex.MatchString(path)
	} else {
		return p.pattern == path
	}
}

func (p *PathPattern) MatchScore(path string) uint8 {
	if p.regex != nil {
		if p.regex.MatchString(path) {
			return 1
		}
	} else if p.pattern == path {
		return 2
	}

	return 0
}

type UrlPattern struct {
	pattern string
	domain  *DomainPattern
	port    string
	path    *PathPattern
	query   map[string]string
}

type UrlSection struct {
	domain string
	port   string
	path   string
	query  map[string]string
}

func NewUrlPattern(pattern string) *UrlPattern {
	u := new(UrlPattern)
	u.pattern = pattern

	s := parseUrlAsPattern(pattern)
	if len(s[0]) > 0 {
		u.domain = NewDomainPattern(s[0])
	}

	u.port = s[1]
	u.path = NewPathPattern(s[2])
	u.query = parseQuery(s[3])
	return u
}

func parseUrlSection(url string) *UrlSection {
	s := parseUrlAsPattern(url)

	u := new(UrlSection)
	u.domain = s[0]
	u.port = s[1]
	u.path = s[2]
	u.query = parseQuery(s[3])
	return u
}

func (p *UrlPattern) Match(url *UrlSection) bool {
	domainMatch := p.domain == nil || p.domain.Match(url.domain)
	portMatch := p.port == url.port
	pathMatch := p.path.Match(url.path)
	queryMatch := matchQuery(p.query, url.query)
	return domainMatch && portMatch && pathMatch && queryMatch
}

func (p *UrlPattern) MatchScore(url *UrlSection) uint32 {
	var domainScore uint8 = 0
	if p.domain != nil {
		domainScore = p.domain.MatchScore(url.domain)
		if domainScore == 0 {
			return 0
		} else {
			domainScore++
		}
	} else {
		domainScore = 1
	}

	if p.port != url.port {
		return 0
	}

	pathScore := p.path.MatchScore(url.path)
	if pathScore == 0 {
		return 0
	}

	queryScore := matchQueryScore(p.query, url.query)
	if queryScore == 0 {
		return 0
	}

	return (uint32(domainScore) << 16) + (uint32(pathScore) << 8) + uint32(queryScore)
}

func parseUrlAsPattern(url string) [4]string {
	if strings.HasPrefix(url, "http://") {
		url = url[len("http://"):]
	}

	head := ""
	s := strings.Index(url, "/")
	if s >= 0 {
		head = url[0:s]
	} else {
		head = url
	}

	host, port, err := net.SplitHostPort(head)
	if err != nil {
		host = head
		port = ""
	} else {
		if port == "80" {
			port = ""
		}
	}

	q := strings.Index(url, "?")
	if s >= 0 && q > s {
		return [4]string{host, port, url[s:q], url[q:]}
	} else if s >= 0 {
		return [4]string{host, port, url[s:], ""}
	} else {
		return [4]string{host, port, "/", ""}
	}
}

func parseQuery(query string) map[string]string {
	m := make(map[string]string)
	if len(query) > 0 && query[0] == '?' {
		query = query[1:]
	}

	kv := strings.Split(query, "&")
	for _, s := range kv {
		if len(s) == 0 {
			continue
		}

		e := strings.Index(s, "=")
		if e == 0 {
			continue
		} else if e < 0 {
			m[s] = ""
		} else {
			m[s[:e]] = s[e+1:]
		}
	}

	return m
}

func matchQuery(pattern, query map[string]string) bool {
	if len(pattern) == 0 {
		return true
	}

	for k, v := range pattern {
		value, ok := query[k]
		if !ok || v != value {
			return false
		}
	}

	return true
}

func matchQueryScore(pattern, query map[string]string) uint8 {
	if len(pattern) == 0 {
		return 1
	}

	for k, v := range pattern {
		value, ok := query[k]
		if !ok || v != value {
			return 0
		}
	}

	if len(pattern) > 255 {
		return 255
	} else {
		return uint8(len(pattern))
	}
}
