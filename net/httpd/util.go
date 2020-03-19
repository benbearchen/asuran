package httpd

import (
	"strings"
)

func MatchPath(path, match string) (string, bool) {
	if path == match {
		return "", true
	}

	m := match
	if !strings.HasSuffix(m, "/") {
		m += "/"
	}

	if strings.HasPrefix(path, m) {
		return path[len(m)-1:], true
	} else {
		return "", false
	}
}

func PopPath(path string) (node string, rest string) {
	for len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	if len(path) == 0 || path == "/" {
		return "", ""
	}

	p := strings.IndexByte(path, '/')
	if p < 0 {
		return path, ""
	} else {
		return path[:p], path[p:]
	}
}

func JoinPath(parent, sub string) string {
	if len(sub) <= 0 {
		return parent
	} else if len(parent) <= 0 {
		return sub
	}

	p := parent[len(parent)-1] == '/'
	q := sub[0] == '/'
	if p && q {
		return parent + sub[1:]
	} else if p != q {
		return parent + sub
	} else {
		return parent + "/" + sub
	}
}
