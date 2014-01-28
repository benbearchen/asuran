package profile

import (
	"testing"
)

func TestDomainPatternRegex(t *testing.T) {
	f := func(p, r string) {
		x := domainPattern2Regex(p)
		if r != x {
			t.Errorf("%s -> %s != %s", p, x, r)
		}
	}

	//f("")
	//f(".")
	//f("*")
	f("domain.com", "^domain\\.com$")
	f("domain.com.", "^domain\\.com$")
	f("domain.*", "^domain\\.[^.]+$")
	f("*.domain.com", "^([^.]+\\.)*domain\\.com$")
	f("cdn.*.domain.com", "^cdn\\.([^.]+\\.)+domain\\.com$")
	f("cdn.*.*.domain.com", "^cdn\\.([^.]+\\.)+([^.]+\\.)+domain\\.com$")
	f("*cdn.domain.com", "^[^.]*cdn\\.domain\\.com$")
	f("abc*cdn.domain.com", "^abc[^.]*cdn\\.domain\\.com$")
	f("*cdn*.domain.com", "^[^.]*cdn[^.]*\\.domain\\.com$")
	f("cdn*.domain.com", "^cdn[^.]*\\.domain\\.com$")
	f("*.cdn.*.domain.com", "^([^.]+\\.)*cdn\\.([^.]+\\.)+domain\\.com$")
}

func TestDomainPattern(t *testing.T) {
	f := func(p, domain string, b bool) {
		if NewDomainPattern(p).Match(domain) != b {
			t.Errorf("%s match %s != %v", p, domain, b)
		}
	}

	//f("")
	//f(".")
	//f("*")
	f("domain.com", "domain.com", true)
	f("domain.com", "domain", false)
	f("domain.com", "domaincom", false)
	f("domain.com", "omaincom", false)
	f("domain.com", "x-domain.com", false)
	f("domain.com", "x.domain.com", false)

	f("domain.*", "domain.com", true)
	f("domain.*", "domain.net", true)
	f("domain.*", "domain.net.com", false)

	f("*.domain.com", "domain.com", true)
	f("*.domain.com", "a.domain.com", true)
	f("*.domain.com", "b.a.domain.com", true)
	f("*.domain.com", "domain", false)
	f("*.domain.com", "omain.com", false)
	f("*.domain.com", "x-domain.com", false)

	f("*.domain.com.", "domain.com", true)
	f("*.domain.com.", "a.domain.com", true)
	f("*.domain.com.", "b.a.domain.com", true)
	f("*.domain.com.", "domain", false)
	f("*.domain.com.", "omain.com", false)
	f("*.domain.com.", "x-domain.com", false)

	f("cdn.*.domain.com", "cdn.a.domain.com", true)
	f("cdn.*.domain.com", "cdn.a.b.domain.com", true)
	f("cdn.*.domain.com", "cdn.a.b.c.domain.com", true)
	f("cdn.*.domain.com", "cn.a.domain.com", false)
	f("cdn.*.domain.com", "cdn..domain.com", false)
	f("cdn.*.domain.com", "cdn.domain.com", false)
	f("cdn.*.domain.com", "x-cdn.domain.com", false)
	f("cdn.*.domain.com", "x.cdn.domain.com", false)
	f("cdn.*.domain.com", "cdn.x-domain.com", false)
	f("cdn.*.domain.com", "x-cdn.x-domain.com", false)
	f("cdn.*.domain.com", "x.cdn.x-domain.com", false)

	f("cdn.*.*.domain.com", "cdn.a.b.domain.com", true)
	f("cdn.*.*.domain.com", "cdn.cdn.b.domain.com", true)
	f("cdn.*.*.domain.com", "cdn.a.cdn.domain.com", true)
	f("cdn.*.*.domain.com", "cdn.a.b.c.domain.com", true)
	f("cdn.*.*.domain.com", "cdn.a.domain.com", false)
	f("cdn.*.*.domain.com", "cn.a.domain.com", false)
	f("cdn.*.*.domain.com", "cdn..domain.com", false)
	f("cdn.*.*.domain.com", "cdn...domain.com", false)
	f("cdn.*.*.domain.com", "cdn.domain.com", false)
	f("cdn.*.*.domain.com", "x-cdn.domain.com", false)
	f("cdn.*.*.domain.com", "x.cdn.domain.com", false)
	f("cdn.*.*.domain.com", "cdn.x-domain.com", false)
	f("cdn.*.*.domain.com", "x-cdn.x-domain.com", false)
	f("cdn.*.*.domain.com", "x.cdn.x-domain.com", false)

	f("*cdn.domain.com", "cdn.domain.com", true)
	f("*cdn.domain.com", "a-cdn.domain.com", true)
	f("*cdn.domain.com", "cn.domain.com", false)
	f("*cdn.domain.com", "cdn-a.domain.com", false)
	f("*cdn.domain.com", "cdn.a.domain.com", false)
	f("*cdn.domain.com", "a.cdn.domain.com", false)
	f("*cdn.domain.com", "cdn.a-domain.com", false)

	f("abc*cdn.domain.com", "abccdn.domain.com", true)
	f("abc*cdn.domain.com", "abc-cdn.domain.com", true)
	f("abc*cdn.domain.com", "abcdn.domain.com", false)
	f("abc*cdn.domain.com", "abc.cdn.domain.com", false)
	f("abc*cdn.domain.com", "a.abc-cdn.domain.com", false)
	f("abc*cdn.domain.com", "abc.abc-cdn.domain.com", false)

	f("*cdn*.domain.com", "cdn.domain.com", true)
	f("*cdn*.domain.com", "x-cdn.domain.com", true)
	f("*cdn*.domain.com", "x-cdn-y.domain.com", true)
	f("*cdn*.domain.com", "cdn-y.domain.com", true)
	f("*cdn*.domain.com", "cdn.cdn.domain.com", false)
	f("*cdn*.domain.com", "cdn.x-domain.com", false)
	f("*cdn*.domain.com", "cdn..domain.com", false)
	f("*cdn*.domain.com", "cn.domain.com", false)

	f("cdn*.domain.com", "cdn.domain.com", true)
	f("cdn*.domain.com", "cdn-x.domain.com", true)
	f("cdn*.domain.com", "cdn.x.domain.com", false)
	f("cdn*.domain.com", "x-cdn.domain.com", false)
	f("cdn*.domain.com", "cdn.xdomain.com", false)
	f("cdn*.domain.com", "domain.com", false)

	f("*.cdn.*.domain.com", "cdn.a.domain.com", true)
	f("*.cdn.*.domain.com", "b.cdn.a.domain.com", true)
	f("*.cdn.*.domain.com", "c.b.cdn.a.domain.com", true)
	f("*.cdn.*.domain.com", "c.b.cdn.a.d.domain.com", true)
	f("*.cdn.*.domain.com", "c.b.cdn.domain.com", false)
	f("*.cdn.*.domain.com", "c.b.x-cdn.d.domain.com", false)
}

func TestPathPatternRegex(t *testing.T) {
	f := func(p, r string) {
		x := pathPattern2Regex(p)
		if r != x {
			t.Errorf("%s -> %s != %s", p, x, r)
		}
	}

	f("", "^/$")
	f("/", "^/$")

	f("a", "^/a$")
	f("/a", "^/a$")
	f("a/", "^/a/$")
	f("/a/", "^/a/$")
	f("a/b", "^/a/b$")
	f("/a/b", "^/a/b$")

	f("*", "^/([^/]+/)*[^/]*$")
	f("/*", "^/([^/]+/)*[^/]*$")
	f("/*/", "^(/[^/]+)+/$")

	f("/*a", "^/[^/]*a$")
	f("/a*", "^/a[^/]*$")
	f("/*a*", "^/[^/]*a[^/]*$")
	f("/*a*", "^/[^/]*a[^/]*$")

	f("/*a/", "^/[^/]*a/$")
	f("/a*/", "^/a[^/]*/$")
	f("/*a*/", "^/[^/]*a[^/]*/$")
	f("/*a*/", "^/[^/]*a[^/]*/$")
}
