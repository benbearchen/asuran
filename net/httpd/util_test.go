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

	check := func(a, b, c string, d bool) {
		p, q := MatchPath(a, b)
		if p != c || q != d {
			t.Errorf("MatchPath(%v, %v) -> (%v, %v) != (%v, %v)", a, b, p, q, c, d)
		}
	}

	check("", "", "", true)
	check("", "a", "", false)
	check("", "a/", "", false)
	check("", "abc", "", false)
	check("", "abc/", "", false)
	check("", "/", "", false)
	check("", "/abc", "", false)
	check("", "/abc/", "", false)
	check("", "/abc/d", "", false)
	check("a", "a", "", true)
	check("a", "ab", "", false)
	check("/", "", "/", true)
	check("/", "a", "", false)
	check("/", "a/", "", false)
	check("/", "abc", "", false)
	check("/", "abc/", "", false)
	check("/", "/", "", true)
	check("/", "/a", "", false)
	check("/", "/a/", "", false)
	check("/", "/abc", "", false)
	check("/", "/abc/", "", false)
	check("/", "/abc/d", "", false)
	check("/a", "", "/a", true)
	check("/a", "a", "", false)
	check("/a", "a/", "", false)
	check("/a", "abc", "", false)
	check("/a", "abc/", "", false)
	check("/a", "/", "/a", true)
	check("/a", "/a", "", true)
	check("/a", "/a/", "", false)
	check("/a", "/abc", "", false)
	check("/a", "/abc/", "", false)
	check("/a", "/abc/d", "", false)
	check("/abc", "", "/abc", true)
	check("/abc", "a", "", false)
	check("/abc", "a/", "", false)
	check("/abc", "abc", "", false)
	check("/abc", "abc/", "", false)
	check("/abc", "/", "/abc", true)
	check("/abc", "/a", "", false)
	check("/abc", "/a/", "", false)
	check("/abc", "/abc", "", true)
	check("/abc", "/abc/", "", false)
	check("/abc", "/abc/d", "", false)
	check("/abc/", "", "/abc/", true)
	check("/abc/", "a", "", false)
	check("/abc/", "a/", "", false)
	check("/abc/", "abc", "", false)
	check("/abc/", "abc/", "", false)
	check("/abc/", "/", "/abc/", true)
	check("/abc/", "/a", "", false)
	check("/abc/", "/a/", "", false)
	check("/abc/", "/abc", "/", true)
	check("/abc/", "/abc/", "", true)
	check("/abc/", "/abc/d", "", false)
	check("/abc/d", "", "/abc/d", true)
	check("/abc/d", "a", "", false)
	check("/abc/d", "a/", "", false)
	check("/abc/d", "abc", "", false)
	check("/abc/d", "abc/", "", false)
	check("/abc/d", "/", "/abc/d", true)
	check("/abc/d", "/a", "", false)
	check("/abc/d", "/a/", "", false)
	check("/abc/d", "/abc", "/d", true)
	check("/abc/d", "/abc/", "/d", true)
	check("/abc/d", "/abc/d", "", true)
}

func TestPopPath(t *testing.T) {
	check := func(a, b, c string) {
		p, q := PopPath(a)
		if p != b || q != c {
			t.Errorf("PopPath(%v) -> (%v, %v) != (%v, %v)", a, p, q, b, c)
		}
	}

	check("", "", "")
	check("a", "a", "")
	check("a/", "a", "/")
	check("a/b", "a", "/b")
	check("a/b/c", "a", "/b/c")
	check("a//", "a", "//")
	check("a//b", "a", "//b")
	check("a//b/", "a", "//b/")
	check("a//b/c", "a", "//b/c")
	check("/", "", "")
	check("/a", "a", "")
	check("/a/", "a", "/")
	check("/a/b", "a", "/b")
	check("/a//", "a", "//")
	check("/a//b", "a", "//b")
	check("//", "", "")
	check("//a", "a", "")
	check("//a/", "a", "/")
	check("//a/b", "a", "/b")
	check("//a//", "a", "//")
	check("//a//b", "a", "//b")
}

func TestJoinPath(t *testing.T) {
	check := func(a, b, c string) {
		d := JoinPath(a, b)
		if d != c {
			t.Errorf("JoinPath('%s', '%s') -> '%s' != '%s'", a, b, d, c)
		}
	}

	check("", "", "")

	check("a", "", "a")
	check("a/", "", "a/")
	check("/a", "", "/a")
	check("/a/", "", "/a/")

	check("", "b", "b")
	check("", "b/", "b/")
	check("", "/b", "/b")
	check("", "/b/", "/b/")

	check("a", "b", "a/b")
	check("a", "b/", "a/b/")
	check("a", "/b", "a/b")
	check("a", "/b/", "a/b/")

	check("a/", "b", "a/b")
	check("a/", "b/", "a/b/")
	check("a/", "/b", "a/b")
	check("a/", "/b/", "a/b/")

	check("/a", "b", "/a/b")
	check("/a", "b/", "/a/b/")
	check("/a", "/b", "/a/b")
	check("/a", "/b/", "/a/b/")

	check("/a/", "b", "/a/b")
	check("/a/", "b/", "/a/b/")
	check("/a/", "/b", "/a/b")
	check("/a/", "/b/", "/a/b/")
}
