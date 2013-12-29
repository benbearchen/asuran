package profile

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

type DelayActType int

const (
	DelayActNone      = iota
	DelayActDelayEach // delay each request in Time seconds
	DelayActDropUntil // drop all request in Time seconds from the first request
)

type DelayAction struct {
	Act  DelayActType
	Time float32 // in seconds
}

func (d *DelayAction) Duration() time.Duration {
	return (time.Duration)(d.Time * 1000000000)
}

func (d *DelayAction) DurationCommand() string {
	if d.Time >= 60*60 {
		return strconv.FormatFloat(float64(d.Time/60/60), 'f', -1, 32) + "h"
	} else if d.Time >= 60 {
		return strconv.FormatFloat(float64(d.Time/60), 'f', -1, 32) + "m"
	} else if d.Time >= 1 {
		return strconv.FormatFloat(float64(d.Time), 'f', -1, 32) + "s"
	} else {
		return strconv.FormatFloat(float64(d.Time*1000), 'f', -1, 32) + "ms"
	}
}

func (d *DelayAction) String() string {
	switch d.Act {
	case DelayActNone:
		return "即时返回"
	case DelayActDelayEach:
		return "固定延时：" + d.DurationCommand()
	case DelayActDropUntil:
		if d.Time == 0 {
			return "丢弃第一次请求"
		} else {
			return "丢弃前 " + d.DurationCommand() + " 内请求"
		}
	default:
		return "delayAct:" + strconv.Itoa(int(d.Act))
	}
}

func MakeEmptyDelay() DelayAction {
	return DelayAction{DelayActNone, 0}
}

func MakeDelay(act DelayActType, delay float32) DelayAction {
	switch act {
	case DelayActNone:
		return DelayAction{act, 0}
	case DelayActDelayEach:
		return DelayAction{act, delay}
	case DelayActDropUntil:
		return DelayAction{act, delay}
	}

	return DelayAction{DelayActNone, 0}
}

type UrlAct int

const (
	UrlActNone = iota
	UrlActCache
	UrlActDrop
)

type UrlProxyAction struct {
	Act              UrlAct
	DropResponseCode int
}

func (action *UrlProxyAction) String() string {
	switch action.Act {
	case UrlActNone:
		return "透明代理"
	case UrlActCache:
		return "缓存"
	case UrlActDrop:
		return "以 " + strconv.Itoa(action.DropResponseCode) + " 丢弃"
	default:
		return "act:" + strconv.Itoa(int(action.Act))
	}
}

func MakeEmptyUrlProxyAction() UrlProxyAction {
	return UrlProxyAction{UrlActNone, 0}
}

type urlAction struct {
	UrlPattern string
	Act        UrlProxyAction
	Delay      DelayAction

	pattern [3]string
}

type UrlOperator interface {
	Action(ip, url string) UrlProxyAction
	Delay(ip, url string) DelayAction
}

type DomainAct int

const (
	DomainActNone = iota
	DomainActBlock
	DomainActRedirect
)

type DomainAction struct {
	Domain string
	Act    DomainAct
	IP     string
}

type DomainOperator interface {
	Action(ip, domain string) DomainAction
}

func (a DomainAct) String() string {
	switch a {
	case DomainActNone:
		return "正常通行"
	case DomainActBlock:
		return "丢弃不返回"
	case DomainActRedirect:
		return "重定向到"
	default:
		return ""
	}
}

func (d *DomainAction) TargetString() string {
	if d.Act != DomainActRedirect {
		return ""
	} else if d.IP == "" {
		return "代理服务器"
	} else {
		return d.IP
	}
}

type HashFileAct int

const (
	HashFileActNone = iota
	HashFileActBlock
	HashFileActSimulate
)

type HashFileAction struct {
	Hash    string
	Act     HashFileAct
	Quality int
}

type HashFileOperator interface {
	Action(hash string) HashFileAct
}

type Profile struct {
	Name    string
	Ip      string
	Owner   string
	Urls    map[string]*urlAction
	Domains map[string]*DomainAction

	lock sync.RWMutex
}

func NewProfile(name, ip, owner string) *Profile {
	p := new(Profile)
	p.Name = name
	p.Ip = ip
	p.Owner = owner
	p.Urls = make(map[string]*urlAction)
	p.Domains = make(map[string]*DomainAction)
	return p
}

func (p *Profile) SetUrlAction(urlPattern string, act UrlAct, responseCode int) {
	p.lock.Lock()
	defer p.lock.Unlock()

	u, ok := p.Urls[urlPattern]
	if ok {
		u.Act = UrlProxyAction{act, responseCode}
	} else {
		u := &urlAction{urlPattern, UrlProxyAction{act, responseCode}, MakeEmptyDelay(), parseUrlAsPattern(urlPattern)}
		p.Urls[urlPattern] = u
	}

	host := getHostOfUrlPattern(urlPattern)
	if len(host) != 0 {
		p.proxyDomainIfNotExists(host)
	}
}

func (p *Profile) UrlAction(url string) UrlProxyAction {
	p.lock.RLock()
	defer p.lock.RUnlock()

	ps := parseUrlAsPattern(url)
	for _, u := range p.Urls {
		if urlPatternMatch(&ps, &u.pattern) {
			return u.Act
		}
	}

	return MakeEmptyUrlProxyAction()
}

func (p *Profile) SetUrlDelay(urlPattern string, act DelayActType, delay float32) {
	p.lock.Lock()
	defer p.lock.Unlock()

	u, ok := p.Urls[urlPattern]
	if ok {
		u.Delay = MakeDelay(act, delay)
	} else {
		u := &urlAction{urlPattern, MakeEmptyUrlProxyAction(), MakeDelay(act, delay), parseUrlAsPattern(urlPattern)}
		p.Urls[urlPattern] = u
	}

	host := getHostOfUrlPattern(urlPattern)
	if len(host) != 0 {
		p.proxyDomainIfNotExists(host)
	}
}

func (p *Profile) UrlDelay(url string) DelayAction {
	p.lock.RLock()
	defer p.lock.RUnlock()

	ps := parseUrlAsPattern(url)
	for _, u := range p.Urls {
		if urlPatternMatch(&ps, &u.pattern) {
			return u.Delay
		}
	}

	return MakeEmptyDelay()
}

func (p *Profile) SetDomainAction(domain string, act DomainAct, targetIP string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	d, ok := p.Domains[domain]
	if ok {
		d.Act = act
		d.IP = targetIP
	} else {
		p.Domains[domain] = &DomainAction{domain, act, targetIP}
	}
}

func (p *Profile) Domain(domain string) DomainAction {
	p.lock.RLock()
	defer p.lock.RUnlock()

	d, ok := p.Domains[domain]
	if ok {
		return *d
	} else {
		return DomainAction{}
	}
}

func (p *Profile) ProxyDomainIfNotExists(domain string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.proxyDomainIfNotExists(domain)
}

func (p *Profile) proxyDomainIfNotExists(domain string) {
	if len(domain) == 0 {
		return
	}

	_, ok := p.Domains[domain]
	if ok {
		return
	}

	p.Domains[domain] = &DomainAction{domain, DomainActRedirect, ""}
}

func (p *Profile) DeleteDomain(domain string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.Domains, domain)
}

func (p *Profile) Delete(urlPattern string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.Urls, urlPattern)
}

func (p *Profile) CloneNew(newName, newIp string) *Profile {
	p.lock.RLock()
	defer p.lock.RUnlock()

	n := NewProfile(newName, newIp, p.Owner)
	for u, url := range p.Urls {
		c := *url
		n.Urls[u] = &c
	}

	for d, domain := range p.Domains {
		c := *domain
		n.Domains[d] = &c
	}

	return n
}

func parseUrlAsPattern(url string) [3]string {
	if strings.HasPrefix(url, "http://") {
		url = url[len("http://"):]
	}

	s := strings.Index(url, "/")
	q := strings.Index(url, "?")
	if s >= 0 && q > s {
		return [3]string{url[0:s], url[s:q], url[q:]}
	} else if s >= 0 {
		return [3]string{url[0:s], url[s:], ""}
	} else {
		return [3]string{url, "", ""}
	}
}

func matchDomain(domain, pattern string) bool {
	if pattern == "" || domain == pattern {
		return true
	}

	// TODO: more pattern
	return false
}

func matchPath(path, pattern string) bool {
	if strings.HasPrefix(path, pattern) {
		return true
	}

	// TODO: more pattern
	return false
}

func parseQuery(query string) map[string]string {
	m := make(map[string]string)
	if len(query) > 0 {
		query = query[1:]
	}

	kv := strings.Split(query, "&")
	for _, s := range kv {
		if len(s) == 0 {
			continue
		}

		e := strings.Index(s, "=")
		if e == 0 {
			continue
		} else if e < 0 {
			m[s] = ""
		} else {
			m[s[:e]] = s[e+1:]
		}
	}

	return m
}

func matchQuery(query, pattern string) bool {
	if pattern == "" {
		return true
	}

	q := parseQuery(query)
	p := parseQuery(pattern)

	for k, v := range p {
		value, ok := q[k]
		if !ok || v != value {
			return false
		}
	}

	return true
}

func urlPatternMatch(url, pattern *[3]string) bool {
	if matchDomain(url[0], pattern[0]) && matchPath(url[1], pattern[1]) && matchQuery(url[2], pattern[2]) {
		return true
	}

	return false
}

func getHostOfUrlPattern(urlPattern string) string {
	p := strings.Index(urlPattern, "/")
	if p <= 0 {
		return ""
	}

	head := urlPattern[0:p]
	p = strings.LastIndex(head, ":")
	if p < 0 {
		return head
	} else {
		return head[0:p]
	}
}
