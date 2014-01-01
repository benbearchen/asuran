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
	f.history.Clear()
}

type cCheckCache struct {
	url       string
	rangeInfo string
	c         chan *cache.UrlCache
}

func (f *Life) CheckCache(url, rangeInfo string) *cache.UrlCache {
	c := make(chan *cache.UrlCache)
	f.c <- cCheckCache{url, rangeInfo, c}
	return <-c
}

func (f *Life) checkCache(url, rangeInfo string) *cache.UrlCache {
	return f.cache.Take(url, rangeInfo)
}

type cLookCache struct {
	url string
	c   chan *cache.UrlCache
}

func (f *Life) LookCache(url string) *cache.UrlCache {
	c := make(chan *cache.UrlCache)
	f.c <- cLookCache{url, c}
	return <-c
}

func (f *Life) lookCache(url string) *cache.UrlCache {
	return f.cache.Look(url)
}

type cListHistory struct {
	url string
	c   chan []*cache.UrlHistory
}

func (f *Life) ListHistory(url string) []*cache.UrlHistory {
	c := make(chan []*cache.UrlHistory)
	f.c <- cListHistory{url, c}
	return <-c
}

func (f *Life) listHistory(url string) []*cache.UrlHistory {
	return f.cache.List(url)
}

type cLookHistoryByID struct {
	id uint32
	c  chan *cache.UrlHistory
}

func (f *Life) LookHistoryByID(id uint32) *cache.UrlHistory {
	c := make(chan *cache.UrlHistory)
	f.c <- cLookHistoryByID{id, c}
	return <-c
}

func (f *Life) lookHistoryByID(id uint32) *cache.UrlHistory {
	return f.cache.History(id)
}

type cSaveContentToCache struct {
	cache *cache.UrlCache
	save  bool
}

func (f *Life) SaveContentToCache(cache *cache.UrlCache, save bool) {
	f.c <- cSaveContentToCache{cache, save}
}

func (f *Life) saveContentToCache(cache *cache.UrlCache, save bool) {
	f.cache.Save(cache, save)
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

type cHistoryEvents struct {
	c chan []*HistoryEvent
}

func (f *Life) HistoryEvents() []*HistoryEvent {
	c := make(chan []*HistoryEvent)
	f.c <- cHistoryEvents{c}
	return <-c
}

func (f *Life) historyEvents() []*HistoryEvent {
	return f.history.Events()
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
			e.c <- f.checkCache(e.url, e.rangeInfo)
		case cLookCache:
			e.c <- f.lookCache(e.url)
		case cListHistory:
			e.c <- f.listHistory(e.url)
		case cLookHistoryByID:
			e.c <- f.lookHistoryByID(e.id)
		case cSaveContentToCache:
			f.saveContentToCache(e.cache, e.save)
		case cLog:
			f.log(e.s)
		case cFormatHistory:
			e.c <- f.formatHistory()
		case cHistoryEvents:
			e.c <- f.historyEvents()
		}
	}
}
