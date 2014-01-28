package httpd

import (
	"strings"
)

func MatchPath(path, match string) (string, bool) {
	if path == match {
		return "", true
	}

	if strings.HasPrefix(path, match+"/") {
		return path[len(match):], true
	} else {
		return "", false
	}
}
