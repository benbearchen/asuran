package cache

import (
	"github.com/benbearchen/asuran/net"

	"net/http"
)

type UrlCache struct {
	Url          string
	Bytes        []byte
	Header       http.Header
	ResponseCode int
	RangeInfo    string
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

type Cache struct {
	contents map[string]UrlCache
}

func NewCache() *Cache {
	c := new(Cache)
	c.contents = make(map[string]UrlCache)

	return c
}

func (c *Cache) Save(cache *UrlCache) {
	c.contents[cache.Url] = *cache
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
	if content, ok := c.contents[url]; ok {
		return &content
	}

	return nil
}

func (c *Cache) Clear() {
	c.contents = make(map[string]UrlCache)
}

func CheckRange(r *http.Request) string {
	range_, ok := r.Header["Range"]
	if ok && len(range_) > 0 {
		return range_[0]
	} else {
		return ""
	}
}
