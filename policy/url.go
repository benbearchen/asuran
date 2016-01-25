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
		hostKeyword,
		removeKeyword,
		deleteKeyword,
	)

	regFactory(new(urlPolicyFactory))

}

type UrlPolicy struct {
	target   string
	set      *SetPolicy
	delays   Policy // drop, delay, timeout
	contents Policy // proxy, cache, map, redirect, rewrite, restore, tcpwrite
	bodys    Policy // delay body, timeout body
	subs     []Policy
	subKeys  map[string]Policy

	def *UrlPolicy
}

func NewDefaultUrlPolicy() *UrlPolicy {
	p, _ := newUrlPolicy([]Policy{}, "")
	return p
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
	} else if len(left) >= 1 {
		return nil, left, fmt.Errorf("unnecessary args: %v", left)
	}

	p, err := newUrlPolicy(subs, target)
	return p, left, err
}

func newUrlPolicy(subs []Policy, target string) (*UrlPolicy, error) {
	u := new(UrlPolicy)
	u.target = target
	u.subs = make([]Policy, 0, 4)
	u.subKeys = make(map[string]Policy)

	for _, p := range subs {
		switch p := p.(type) {
		case *SetPolicy:
			u.set = p
		case *DelayPolicy, *TimeoutPolicy, *DropPolicy:
			if p.(baseDelayInterface).Body() {
				if u.bodys != nil {
					return nil, fmt.Errorf(`conflict keyword: "%s" vs "%s"`, u.bodys.Command(), p.Command())
				} else {
					u.bodys = p
				}
			} else {
				if u.delays != nil {
					return nil, fmt.Errorf(`conflict keyword: "%s" vs "%s"`, u.delays.Command(), p.Command())
				} else {
					u.delays = p
				}
			}
		case *ProxyPolicy, *CachePolicy, *MapPolicy, *RedirectPolicy, *RewritePolicy, *RestorePolicy, *TcpwritePolicy:
			if u.contents != nil {
				return nil, fmt.Errorf(`conflict keyword: "%s" vs "%s"`, u.contents.Command(), p.Command())
			} else {
				u.contents = p
			}
		default:
			u.subs = append(u.subs, p)
			u.subKeys[p.Keyword()] = p
		}
	}

	return u, nil
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
	if u.delays != nil {
		s = append(s, u.delays.Command())
	}

	if u.contents != nil {
		s = append(s, u.contents.Command())
	}

	if u.bodys != nil {
		s = append(s, u.bodys.Command())
	}

	for _, p := range u.subs {
		s = append(s, p.Command())
	}

	s = append(s, u.target)
	return strings.Join(s, " ")
}

func (u *UrlPolicy) Comment() string {
	c := make([]string, 0, len(u.subs))
	if u.delays != nil {
		c = append(c, u.delays.Comment())
	}

	if u.contents != nil {
		c = append(c, u.contents.Comment())
	}

	if u.bodys != nil {
		c = append(c, u.bodys.Comment())
	}

	for _, p := range u.subs {
		c = append(c, p.Comment())
	}

	return strings.Join(c, "；")
}

func (u *UrlPolicy) OtherComment() string {
	c := make([]string, 0, len(u.subs))
	if u.bodys != nil {
		c = append(c, u.bodys.Comment())
	}

	for _, p := range u.subs {
		c = append(c, p.Comment())
	}

	return strings.Join(c, "；")
}

func (u *UrlPolicy) Set() bool {
	return u.set != nil && u.set.Value()
}

func (u *UrlPolicy) Update(p Policy) error {
	switch p := p.(type) {
	case *UrlPolicy:
		return u.update(p)
	case *DelayPolicy, *TimeoutPolicy, *DropPolicy:
		if p.(baseDelayInterface).Body() {
			u.bodys = p
		} else {
			u.delays = p
		}
	case *ProxyPolicy, *CachePolicy, *MapPolicy, *RedirectPolicy, *RewritePolicy, *RestorePolicy, *TcpwritePolicy:
		u.contents = p
	case *StatusPolicy, *SpeedPolicy, *Dont302Policy, *Disable304Policy, *ContentTypePolicy:
		for i, s := range u.subs {
			if s.Keyword() == p.Keyword() {
				u.subs[i] = p
				u.subKeys[p.Keyword()] = p
				return nil
			}
		}

		u.subs = append(u.subs, p)
		u.subKeys[p.Keyword()] = p
	default:
		return fmt.Errorf("unmatch policy to url: %s", p.Command())
	}

	return nil
}

func (u *UrlPolicy) update(p *UrlPolicy) error {
	if p.Set() {
		u.delays = p.delays
		u.contents = p.contents
		u.bodys = p.bodys
		u.subs = p.subs
		u.subKeys = p.subKeys
	} else {
		if p.delays != nil {
			u.delays = p.delays
		}

		if p.contents != nil {
			u.contents = p.contents
		}

		if p.bodys != nil {
			u.bodys = p.bodys
		}

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

func (u *UrlPolicy) Target() string {
	return u.target
}

func (u *UrlPolicy) Speed() *SpeedPolicy {
	p := u.subKeyDef(speedKeyword)
	if p != nil {
		p, ok := p.(*SpeedPolicy)
		if ok {
			return p
		}
	}

	return nil
}

func (u *UrlPolicy) Status() *StatusPolicy {
	p := u.subKeyDef(statusKeyword)
	if p != nil {
		p, ok := p.(*StatusPolicy)
		if ok {
			return p
		}
	}

	return nil
}

func (u *UrlPolicy) Dont302() bool {
	p := u.subKeyDef(dont302Keyword)
	if p != nil {
		d, ok := p.(*Dont302Policy)
		if ok {
			return d.Value()
		}
	}

	return false
}

func (u *UrlPolicy) Disable304() bool {
	p := u.subKeyDef(disable304Keyword)
	if p != nil {
		d, ok := p.(*Disable304Policy)
		if ok {
			return d.Value()
		}
	}

	return false
}

func (u *UrlPolicy) ContentType() string {
	p := u.subKeyDef(contentTypeKeyword)
	if p != nil {
		c, ok := p.(*ContentTypePolicy)
		if ok {
			return c.Value()
		}
	}

	return ContentTypeActDefault
}

func (u *UrlPolicy) Host() *HostPolicy {
	p := u.subKeyDef(hostKeyword)
	if p != nil {
		c, ok := p.(*HostPolicy)
		if ok {
			return c
		}
	}

	return nil
}

func (u *UrlPolicy) Delete() bool {
	_, ok := u.subKeys[deleteKeyword]
	return ok
}

func (u *UrlPolicy) DelayPolicy() Policy {
	if u.delays != nil {
		return u.delays
	} else if u.def != nil {
		return u.def.DelayPolicy()
	} else {
		return nil
	}
}

func (u *UrlPolicy) DelayComment() string {
	if u.delays != nil {
		return u.delays.Comment()
	} else if u.def != nil {
		return "[" + u.def.DelayComment() + "]"
	} else {
		return "即时返回"
	}
}

func (u *UrlPolicy) ContentPolicy() Policy {
	if u.contents != nil {
		return u.contents
	} else if u.def != nil {
		return u.def.ContentPolicy()
	} else {
		return nil
	}
}

func (u *UrlPolicy) ContentComment() string {
	if u.contents != nil {
		return u.contents.Comment()
	} else if u.def != nil {
		return "[" + u.def.ContentComment() + "]"
	} else {
		return "透明代理"
	}
}

func (u *UrlPolicy) BodyPolicy() Policy {
	if u.bodys != nil {
		return u.bodys
	} else if u.def != nil {
		return u.def.BodyPolicy()
	} else {
		return nil
	}
}

func (u *UrlPolicy) Def(def *UrlPolicy) {
	u.def = def
}

func (u *UrlPolicy) subKeyDef(key string) Policy {
	p, _ := u.subKeys[key]
	if p != nil {
		return p
	}

	if u.def != nil {
		return u.def.subKeyDef(key)
	}

	return nil
}
