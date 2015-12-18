package policy

import (
	"fmt"
	"strings"
)

const urlKeyword = "url"

var urlSubKeys sub

func init() {
	urlSubKeys.init(
		defaultKeyword,
		proxyKeyword,
		statusKeyword,
		speedKeyword,
		timeoutKeyword,
		dropKeyword,
	)

	regFactory(new(urlPolicyFactory))

}

type UrlPolicy struct {
	target  string
	subs    []Policy
	subKeys map[string]Policy
}

type urlPolicyFactory struct {
}

func (*urlPolicyFactory) Keyword() string {
	return urlKeyword
}

func (*urlPolicyFactory) Build(args []string) (Policy, []string, error) {
	left := args
	subs := make([]Policy, 0)
	for len(left) > 0 {
		keyword, rest := left[0], left[1:]
		if !urlSubKeys.isSubKey(keyword) {
			if len(rest) != 0 {
				return nil, rest, fmt.Errorf(`rest args: %v is unnecessary`, rest)
			} else {
				break
			}
		}

		sub, rest, err := urlSubKeys.makeSub(keyword, rest)
		if err != nil {
			return nil, rest, err
		} else {
			subs = append(subs, sub)
			left = rest
		}
	}

	if len(left) == 0 {
		return nil, left, fmt.Errorf("missing target url or pattern")
	}

	return newUrlPolicy(subs, left[0]), nil, nil
}

func newUrlPolicy(subs []Policy, target string) *UrlPolicy {
	u := new(UrlPolicy)
	u.target = target
	u.subs = subs
	u.subKeys = make(map[string]Policy)
	for _, p := range subs {
		u.subKeys[p.Keyword()] = p
	}

	return u
}

func (u *UrlPolicy) Keyword() string {
	return urlKeyword
}

func (u *UrlPolicy) Command() string {
	s := make([]string, 0, 2+len(u.subs))
	s = append(s, urlKeyword)
	for _, p := range u.subs {
		s = append(s, p.Command())
	}

	s = append(s, u.target)
	return strings.Join(s, " ")
}

func (u *UrlPolicy) Comment() string {
	c := make([]string, 0, len(u.subs))
	for _, p := range u.subs {
		c = append(c, p.Comment())
	}

	return strings.Join(c, "；")
}

func (u *UrlPolicy) Speed() *SpeedPolicy {
	p, _ := u.subKeys[speedKeyword]
	if p != nil {
		p, ok := p.(*SpeedPolicy)
		if ok {
			return p
		}
	}

	return nil
}

func (u *UrlPolicy) Status() *StatusPolicy {
	p, _ := u.subKeys[statusKeyword]
	if p != nil {
		p, ok := p.(*StatusPolicy)
		if ok {
			return p
		}
	}

	return nil
}
