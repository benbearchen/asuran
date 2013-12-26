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
}

func NewUrlCache(url string, resp *net.HttpResponse, content []byte) *UrlCache {
	return &UrlCache{url, content, resp.Header(), resp.ResponseCode()}
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

func (c *Cache) Take(url string) *UrlCache {
	if content, ok := c.contents[url]; ok {
		return &content
	} else {
		return nil
	}
}

func (c *Cache) Clear() {
	c.contents = make(map[string]UrlCache)
}
