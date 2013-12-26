package life

import (
	"github.com/benbearchen/asuran/profile"
	"github.com/benbearchen/asuran/web/proxy/cache"

	"time"
)

type Event struct {
	Time time.Time
}

type UrlEvent struct {
	Event

	DelayAct profile.DelayActType
	ProxyAct profile.UrlProxyAction
}

type UrlState struct {
	Url        string
	CreateTime time.Time
	BeginTime  time.Time
	Events     []UrlEvent
}

func (u *UrlState) DropUntil(duration time.Duration) bool {
	now := time.Now()
	if u.BeginTime.IsZero() {
		u.BeginTime = time.Now()
	}

	d := now.Sub(u.BeginTime)
	return d < duration
}

type DomainEvent struct {
	Event

	Act profile.DomainAct
	IP  string
}

type DomainState struct {
	Domain     string
	CreateTime time.Time
	BeginTime  time.Time
	Events     []DomainEvent
}

type Life struct {
	IP      string
	urls    map[string]*UrlState
	domains map[string]*DomainState
	cache   *cache.Cache
	history *History

	c chan interface{}
}

func NewLife(ip string) *Life {
	f := Life{}
	f.IP = ip
	f.urls = make(map[string]*UrlState)
	f.domains = make(map[string]*DomainState)
	f.cache = cache.NewCache()
	f.history = NewHistory()

	f.c = make(chan interface{})
	go func() {
		f.work()
	}()

	return &f
}

type cOpenUrl struct {
	url string
	c   chan *UrlState
}

func (f *Life) OpenUrl(url string) *UrlState {
	c := make(chan *UrlState)
	f.c <- cOpenUrl{url, c}
	return <-c
}

func (f *Life) openUrl(url string) *UrlState {
	u, ok := f.urls[url]
	if !ok {
		now := time.Now()
		u = &UrlState{url, now, now, make([]UrlEvent, 0)}
		f.urls[url] = u
	} else {
		if u.BeginTime.IsZero() {
			u.BeginTime = time.Now()
		}
	}

	return u
}

type cOpenDomain struct {
	domain string
	c      chan *DomainState
}

func (f *Life) OpenDomain(domain string) *DomainState {
	c := make(chan *DomainState)
	f.c <- cOpenDomain{domain, c}
	return <-c
}

func (f *Life) openDomain(domain string) *DomainState {
	d, ok := f.domains[domain]
	if !ok {
		now := time.Now()
		d = &DomainState{domain, now, now, make([]DomainEvent, 0)}
		f.domains[domain] = d
	} else {
		if d.BeginTime.IsZero() {
			d.BeginTime = time.Now()
		}
	}

	return d
}

type cRestart struct {
}

func (f *Life) Restart() {
	f.c <- cRestart{}
}

func (f *Life) restart() {
	for _, u := range f.urls {
		u.BeginTime = time.Time{}
	}

	for _, d := range f.domains {
		d.BeginTime = time.Time{}
	}

	f.cache.Clear()
}

type cCheckCache struct {
	url string
	c   chan *cache.UrlCache
}

func (f *Life) CheckCache(url string) *cache.UrlCache {
	c := make(chan *cache.UrlCache)
	f.c <- cCheckCache{url, c}
	return <-c
}

func (f *Life) checkCache(url string) *cache.UrlCache {
	return f.cache.Take(url)
}

type cSaveContentToCache struct {
	url     string
	content string
}

func (f *Life) SaveContentToCache(url string, content string) {
	f.c <- cSaveContentToCache{url, content}
}

func (f *Life) saveContentToCache(url string, content string) {
	f.cache.Save(url, []byte(content))
}

type cLog struct {
	s string
}

func (f *Life) Log(s string) {
	f.c <- cLog{s}
}

func (f *Life) log(s string) {
	f.history.Log(StringHistory(s))
}

type cFormatHistory struct {
	c chan string
}

func (f *Life) FormatHistory() string {
	c := make(chan string)
	f.c <- cFormatHistory{c}
	return <-c
}

func (f *Life) formatHistory() string {
	return f.history.Format()
}

func (f *Life) work() {
	for {
		e, ok := <-f.c
		if !ok {
			return
		}

		switch e := e.(type) {
		case cOpenUrl:
			e.c <- f.openUrl(e.url)
		case cOpenDomain:
			e.c <- f.openDomain(e.domain)
		case cRestart:
			f.restart()
		case cCheckCache:
			e.c <- f.checkCache(e.url)
		case cSaveContentToCache:
			f.saveContentToCache(e.url, e.content)
		case cLog:
			f.log(e.s)
		case cFormatHistory:
			e.c <- f.formatHistory()
		}
	}
}
