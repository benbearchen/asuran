package life

import (
	"github.com/benbearchen/asuran/web/proxy/cache"

	"time"
)

type Event struct {
	Time time.Time
}

type UrlEvent struct {
	Event
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

	Act string
	IP  string
}

type DomainState struct {
	Domain     string
	CreateTime time.Time
	BeginTime  time.Time
	Events     []DomainEvent
}

type Life struct {
	IP         string
	CreateTime time.Time
	VisitTime  time.Time
	urls       map[string]*UrlState
	domains    map[string]*DomainState
	cache      *cache.Cache
	history    *History
	watching   []cWatchHistory

	c chan interface{}
}

func NewLife(ip string) *Life {
	f := Life{}
	f.IP = ip
	f.CreateTime = time.Now()
	f.VisitTime = f.CreateTime
	f.urls = make(map[string]*UrlState)
	f.domains = make(map[string]*DomainState)
	f.cache = cache.NewCache()
	f.history = NewHistory()
	f.watching = make([]cWatchHistory, 0)

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
	f.clearHistory()
	f.VisitTime = time.Time{}
}

type cClearHistory struct {
}

func (f *Life) ClearHistory() {
	f.c <- cClearHistory{}
}

func (f *Life) clearHistory() {
	f.history.Clear()

	go func(w []cWatchHistory) {
		for _, e := range w {
			e.c <- true
		}
	}(f.watching)

	f.watching = make([]cWatchHistory, 0)
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
	c     chan uint32
}

func (f *Life) SaveContentToCache(cache *cache.UrlCache, save bool) uint32 {
	c := make(chan uint32)
	f.c <- cSaveContentToCache{cache, save, c}
	return <-c
}

func (f *Life) saveContentToCache(cache *cache.UrlCache, save bool) uint32 {
	return f.cache.Save(cache, save)
}

type cLog struct {
	s string
}

func (f *Life) Log(s string) {
	f.c <- cLog{s}
}

func (f *Life) log(s string) {
	event := StringHistory(s)
	f.history.Log(event)

	go func(w []cWatchHistory) {
		for _, e := range w {
			e.c <- []*HistoryEvent{&event}
		}
	}(f.watching)

	f.watching = make([]cWatchHistory, 0)
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

type cWatchHistory struct {
	c chan interface{}
	t time.Time
}

func (f *Life) WatchHistory(t time.Time) chan interface{} {
	c := make(chan interface{})
	f.c <- cWatchHistory{c, t}
	return c
}

func (f *Life) watchHistory(e cWatchHistory) {
	events := f.history.EventsAfter(e.t)
	if len(events) > 0 {
		go func() {
			e.c <- events
		}()
	} else {
		f.watching = append(f.watching, e)
	}
}

type cStopWatchHistory struct {
	c chan interface{}
}

func (f *Life) StopWatchHistory(c chan interface{}) {
	f.c <- cStopWatchHistory{c}
}

func (f *Life) stopWatchHistory(c chan interface{}) {
	for i, e := range f.watching {
		if e.c == c {
			w := make([]cWatchHistory, len(f.watching)-1)
			copy(w[:i], f.watching[:i])
			copy(w[i:], f.watching[i+1:])
			f.watching = w
			return
		}
	}
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
		case cClearHistory:
			f.clearHistory()
		case cCheckCache:
			e.c <- f.checkCache(e.url, e.rangeInfo)
		case cLookCache:
			e.c <- f.lookCache(e.url)
		case cListHistory:
			e.c <- f.listHistory(e.url)
		case cLookHistoryByID:
			e.c <- f.lookHistoryByID(e.id)
		case cSaveContentToCache:
			e.c <- f.saveContentToCache(e.cache, e.save)
		case cLog:
			f.log(e.s)
		case cFormatHistory:
			e.c <- f.formatHistory()
		case cHistoryEvents:
			e.c <- f.historyEvents()
		case cWatchHistory:
			f.watchHistory(e)
		}
	}
}

func (f *Life) visit() {
	f.VisitTime = time.Now()
}

func (f *Life) isIdle(d time.Duration) bool {
	if f.VisitTime.IsZero() {
		return false
	}

	return time.Now().Sub(f.VisitTime) > d
}
