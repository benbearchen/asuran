package policy

import (
	"fmt"
	"regexp"
	"strings"
)

type Replacer struct {
	search  *regexp.Regexp
	replace string
}

func NewReplacer(pattern string) (*Replacer, error) {
	search, replace, err := splitSearchReplace(pattern)
	if err != nil {
		return nil, err
	}

	s, err := regexp.Compile(search)
	if err != nil {
		return nil, err
	}

	return &Replacer{s, replace}, nil
}

func (r *Replacer) Replace(source string) string {
	return r.search.ReplaceAllString(source, r.replace)
}

func splitSearchReplace(pattern string) (search string, replace string, err error) {
	ss := strings.Split(pattern, "/")
	if len(ss) < 4 {
		return "", "", fmt.Errorf(`replace pattern "%s" should like "/.../.../"`, pattern)
	}

	return ss[1], ss[2], nil
}
