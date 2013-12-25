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
}

func NewLife(ip string) *Life {
	f := Life{}
	f.IP = ip
	f.urls = make(map[string]*UrlState)
	f.domains = make(map[string]*DomainState)
	f.cache = cache.NewCache()
	f.history = NewHistory()

	return &f
}

func (f *Life) OpenUrl(url string) *UrlState {
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

func (f *Life) OpenDomain(domain string) *DomainState {
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

func (f *Life) Restart() {
	for _, u := range f.urls {
		u.BeginTime = time.Time{}
	}

	for _, d := range f.domains {
		d.BeginTime = time.Time{}
	}

	f.cache.Clear()
}

func (f *Life) CheckCache(url string) ([]byte, bool) {
	return f.cache.Take(url)
}

func (f *Life) SaveContentToCache(url string, content string) {
	f.cache.Save(url, []byte(content))
}

func (f *Life) Log(s string) {
	f.history.Log(StringHistory(s))
}

func (f *Life) FormatHistory() string {
	return f.history.Format()
}
