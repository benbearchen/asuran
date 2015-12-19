package policy

import (
	"fmt"
	"strings"
)

const urlKeyword = "url"

var urlSubKeys sub

func init() {
	urlSubKeys.init(
		setKeyword,
		updateKeyword,
		dropKeyword,
		delayKeyword,
		timeoutKeyword,
		proxyKeyword,
		cacheKeyword,
		statusKeyword,
		mapKeyword,
		redirectKeyword,
		rewriteKeyword,
		restoreKeyword,
		tcpwriteKeyword,
		speedKeyword,
		dont302Keyword,
		do302Keyword,
		disable304Keyword,
		allow304Keyword,
		contentTypeKeyword,
		deleteKeyword,
	)

	regFactory(new(urlPolicyFactory))

}

type UrlPolicy struct {
	target  string
	set     *SetPolicy
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
			break
		}

		sub, rest, err := urlSubKeys.makeSub(keyword, rest)
		if err != nil {
			return nil, rest, err
		} else {
			subs = append(subs, sub)
			left = rest
		}
	}

	target := ""
	if len(left) != 0 {
		target, left = left[0], left[1:]
	}

	if len(left) == 0 {
		left = nil
	}

	return newUrlPolicy(subs, target), left, nil
}

func newUrlPolicy(subs []Policy, target string) *UrlPolicy {
	u := new(UrlPolicy)
	u.target = target
	u.subs = subs
	u.subKeys = make(map[string]Policy)
	for _, p := range subs {
		switch p := p.(type) {
		case *SetPolicy:
			u.set = p
		default:
			u.subKeys[p.Keyword()] = p
		}
	}

	return u
}

func FactoryUrl(cmd string) (*UrlPolicy, error) {
	p, err := Factory(cmd)
	if err != nil {
		return nil, err
	}

	u, ok := p.(*UrlPolicy)
	if ok {
		return u, nil
	} else {
		return nil, fmt.Errorf(`not a URL policy: "%s"`, cmd)
	}
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

func (u *UrlPolicy) Set() bool {
	return u.set != nil && u.set.Value()
}

func (u *UrlPolicy) Update(p Policy) error {
	if u.Keyword() != p.Keyword() {
		return fmt.Errorf("unmatch keywrod: %s vs %s", u.Keyword(), p.Keyword())
	}

	switch p := p.(type) {
	case *UrlPolicy:
		return u.update(p)
	default:
		return fmt.Errorf("unmatch policy")
	}

	return nil
}

func (u *UrlPolicy) update(p *UrlPolicy) error {
	if p.Set() {
		u.subs = p.subs
		u.subKeys = p.subKeys
	} else {
		for i, s := range u.subs {
			key := s.Keyword()
			sub, ok := p.subKeys[key]
			if ok {
				u.subs[i] = sub
				u.subKeys[key] = sub
			}
		}

		for key, sub := range p.subKeys {
			if _, ok := u.subKeys[key]; !ok {
				u.subs = append(u.subs, sub)
				u.subKeys[key] = sub
			}
		}
	}

	return nil
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

func (u *UrlPolicy) Dont302() bool {
	p, _ := u.subKeys[dont302Keyword]
	if p != nil {
		d, ok := p.(*Dont302Policy)
		if ok {
			return d.Value()
		}
	}

	return false
}

func (u *UrlPolicy) Disable304() bool {
	p, _ := u.subKeys[disable304Keyword]
	if p != nil {
		d, ok := p.(*Disable304Policy)
		if ok {
			return d.Value()
		}
	}

	return false
}
