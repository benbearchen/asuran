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
	pattern    *UrlPattern
	Act        UrlProxyAction
	Delay      DelayAction
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
	Domain  string
	pattern *DomainPattern
	Act     DomainAct
	IP      string
}

type DomainOperator interface {
	Action(ip, domain string) *DomainAction
}

func NewDomainAction(domain string, act DomainAct, ip string) *DomainAction {
	return &DomainAction{domain, nil, act, ip}
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
		u := &urlAction{urlPattern, NewUrlPattern(urlPattern), UrlProxyAction{act, responseCode}, MakeEmptyDelay()}
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

	us := parseUrlSection(url)
	for _, u := range p.Urls {
		if u.pattern.Match(us) {
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
		u := &urlAction{urlPattern, NewUrlPattern(urlPattern), MakeEmptyUrlProxyAction(), MakeDelay(act, delay)}
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

	us := parseUrlSection(url)
	for _, u := range p.Urls {
		if u.pattern.Match(us) {
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
		p.Domains[domain] = &DomainAction{domain, NewDomainPattern(domain), act, targetIP}
	}
}

func (p *Profile) Domain(domain string) *DomainAction {
	p.lock.RLock()
	defer p.lock.RUnlock()

	d, ok := p.Domains[domain]
	if ok {
		return d
	}

	for _, d := range p.Domains {
		if d.pattern != nil && d.pattern.Match(domain) {
			return d
		}
	}

	return nil
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

	p.Domains[domain] = &DomainAction{domain, NewDomainPattern(domain), DomainActRedirect, ""}
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
