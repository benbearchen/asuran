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
