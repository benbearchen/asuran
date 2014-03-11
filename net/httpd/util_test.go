package httpd

import "testing"

func TestMatchPath(t *testing.T) {
	f := func(p, q string, r, s bool) {
		if p != q || r != s {
			t.Errorf("%s != %s || %v != %v", p, q, r, s)
		}
	}

	p, r := MatchPath("/profile", "/")
	f(p, "/profile", r, true)

	p, r = MatchPath("/profile", "/profile")
	f(p, "", r, true)

	p, r = MatchPath("/profile/1", "/profile")
	f(p, "/1", r, true)

	p, r = MatchPath("/profile1", "/profile")
	f(p, "", r, false)

	p, r = MatchPath("/profile", "/profile/")
	f(p, "", r, false)

	p, r = MatchPath("/profile/1", "/profile/")
	f(p, "/1", r, true)

	p, r = MatchPath("/profile1", "/profile/")
	f(p, "", r, false)
}
