package cache

import (
	"github.com/benbearchen/asuran/net"

	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

type UrlCache struct {
	Url          string
	Bytes        []byte
	Header       http.Header
	ResponseCode int
	RangeInfo    string
}

type UrlHistory struct {
	UrlCache

	ID   uint32
	Time time.Time
}

func NewUrlCache(url string, resp *net.HttpResponse, content []byte, rangeInfo string) *UrlCache {
	return &UrlCache{url, content, resp.Header(), resp.ResponseCode(), rangeInfo}
}

func (c *UrlCache) Response(w http.ResponseWriter) {
	h := w.Header()
	for k, v := range c.Header {
		h[k] = v
	}

	w.WriteHeader(c.ResponseCode)
	w.Write(c.Bytes)
}

func (c *UrlCache) Detail(w http.ResponseWriter) {
	t := "URL: " + c.Url + "\n"
	if len(c.RangeInfo) > 0 {
		t += "Request Range: " + c.RangeInfo + "\n"
	}

	t += "ResponseCode: " + strconv.Itoa(c.ResponseCode) + "\n"
	t += "Headers: {{{\n"
	for k, v := range c.Header {
		for _, v := range v {
			t += k + ": " + v + "\n"
		}
	}

	t += "}}}\n"

	t += "Content-Length: " + strconv.Itoa(len(c.Bytes))

	fmt.Fprintf(w, t)
}

type Cache struct {
	contents map[string]UrlCache
	urlIds   map[string][]uint32
	id       uint32
	indexes  []*UrlHistory
}

func NewCache() *Cache {
	c := new(Cache)
	c.Clear()

	return c
}

func (c *Cache) historyID() uint32 {
	return atomic.AddUint32(&c.id, 1)
}

func (c *Cache) Save(cache *UrlCache, save bool) {
	if save {
		c.contents[cache.Url] = *cache
	}

	id := c.historyID()
	ids, ok := c.urlIds[cache.Url]
	if !ok {
		ids = make([]uint32, 0)
	}

	c.urlIds[cache.Url] = append(ids, id)
	c.indexes = append(c.indexes, &UrlHistory{*cache, id, time.Now()})
}

func (c *Cache) Take(url, rangeInfo string) *UrlCache {
	if content, ok := c.contents[url]; ok {
		if content.RangeInfo == rangeInfo {
			return &content
		}
	}

	return nil
}

func (c *Cache) Look(url string) *UrlCache {
	h := c.LastHistory(url)
	if h != nil {
		return &h.UrlCache
	}

	return nil
}

func (c *Cache) List(url string) []*UrlHistory {
	h := make([]*UrlHistory, 0)
	if ids, ok := c.urlIds[url]; ok {
		for _, id := range ids {
			history := c.History(id)
			if history != nil {
				h = append(h, history)
			}
		}
	}

	return h
}

func (c *Cache) LastHistory(url string) *UrlHistory {
	if ids, ok := c.urlIds[url]; ok {
		if len(ids) > 0 {
			return c.History(ids[len(ids)-1])
		}
	}

	return nil
}

func (c *Cache) History(id uint32) *UrlHistory {
	if uint32(len(c.indexes)) < id {
		return nil
	}

	return c.indexes[id-1]
}

func (c *Cache) Clear() {
	c.contents = make(map[string]UrlCache)
	c.urlIds = make(map[string][]uint32)
	c.id = 0
	c.indexes = make([]*UrlHistory, 0, 20)
}

func CheckRange(r *http.Request) string {
	range_, ok := r.Header["Range"]
	if ok && len(range_) > 0 {
		return range_[0]
	} else {
		return ""
	}
}
