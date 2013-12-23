package life

import (
	"github.com/benbearchen/asuran/profile"

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
	Urls    map[string]*UrlState
	Domains map[string]*DomainState
}

func NewLife(ip string) *Life {
	f := Life{}
	f.IP = ip
	f.Urls = make(map[string]*UrlState)
	f.Domains = make(map[string]*DomainState)

	return &f
}

func (f *Life) OpenUrl(url string) *UrlState {
	u, ok := f.Urls[url]
	if !ok {
		now := time.Now()
		u = &UrlState{url, now, now, make([]UrlEvent, 0)}
		f.Urls[url] = u
	} else {
		if u.BeginTime.IsZero() {
			u.BeginTime = time.Now()
		}
	}

	return u
}

func (f *Life) OpenDomain(domain string) *DomainState {
	d, ok := f.Domains[domain]
	if !ok {
		now := time.Now()
		d = &DomainState{domain, now, now, make([]DomainEvent, 0)}
		f.Domains[domain] = d
	} else {
		if d.BeginTime.IsZero() {
			d.BeginTime = time.Now()
		}
	}

	return d
}

func (f *Life) Restart() {
	for _, u := range f.Urls {
		u.BeginTime = time.Time{}
	}

	for _, d := range f.Domains {
		d.BeginTime = time.Time{}
	}
}
