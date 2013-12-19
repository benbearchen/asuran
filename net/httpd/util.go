package httpd

import (
	"strings"
)

func MatchPath(path, match string) (string, bool) {
	if strings.HasPrefix(path, match) {
		return path[len(match):], true
	} else {
		return "", false
	}
}
