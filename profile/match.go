package profile

import (
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
			rep := "+"
			if len(r) == 0 {
				rep = "*"
			}

			r += "([^.]+\\.)" + rep
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

	r = "^" + r + "$"
	return r
}

func (d *DomainPattern) Match(domain string) bool {
	if d.regex != nil {
		return d.regex.MatchString(domain)
	} else {
		return d.pattern == domain
	}
}

type UrlPattern struct {
	pattern string
	domain  *DomainPattern
	path    string
	query   map[string]string
}

type UrlSection struct {
	domain string
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

	u.path = s[1]
	u.query = parseQuery(s[2])
	return u
}

func parseUrlSection(url string) *UrlSection {
	s := parseUrlAsPattern(url)

	u := new(UrlSection)
	u.domain = s[0]
	u.path = s[1]
	u.query = parseQuery(s[2])
	return u
}

func (p *UrlPattern) Match(url *UrlSection) bool {
	domainMatch := p.domain == nil || p.domain.Match(url.domain)
	pathMatch := matchPath(p.path, url.path)
	queryMatch := matchQuery(p.query, url.query)
	return domainMatch && pathMatch && queryMatch
}

func parseUrlAsPattern(url string) [3]string {
	if strings.HasPrefix(url, "http://") {
		url = url[len("http://"):]
	}

	s := strings.Index(url, "/")
	q := strings.Index(url, "?")
	if s >= 0 && q > s {
		return [3]string{url[0:s], url[s:q], url[q:]}
	} else if s >= 0 {
		return [3]string{url[0:s], url[s:], ""}
	} else {
		return [3]string{url, "", ""}
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

func matchPath(pattern, path string) bool {
	if strings.HasPrefix(path, pattern) {
		return true
	}

	// TODO: more pattern
	return false
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