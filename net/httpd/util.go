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
		return path[:p], path[p+1:]
	}
}
