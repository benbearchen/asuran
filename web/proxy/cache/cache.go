package cache

import (
	"github.com/benbearchen/asuran/net"

	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"
)

type UrlCache struct {
	Time           time.Time
	Duration       time.Duration
	Url            string
	Method         string
	RequestHeader  http.Header
	PostBody       []byte
	ContentSource  string
	Bytes          []byte
	ResponseHeader http.Header
	ResponseCode   int
	RangeInfo      string
	Error          error
}

type UrlHistory struct {
	UrlCache

	ID uint32
}

func NewUrlCache(url string, r *http.Request, postBody []byte, resp *net.HttpResponse, contentSource string, content []byte, rangeInfo string, start, end time.Time, err error) *UrlCache {
	var respHeader http.Header
	respResponseCode := -1
	if resp != nil {
		respHeader = resp.Header()
		respResponseCode = resp.ResponseCode()
	}

	return &UrlCache{start, end.Sub(start), url, r.Method, r.Header, postBody, contentSource, content, respHeader, respResponseCode, rangeInfo, err}
}

func (c *UrlCache) Response(w http.ResponseWriter, wrap io.Writer) {
	h := w.Header()
	for k, v := range c.ResponseHeader {
		h[k] = v
	}

	w.WriteHeader(c.ResponseCode)
	if wrap == nil {
		wrap = w
	}

	wrap.Write(c.Bytes)
}

func (c *UrlCache) Detail(w http.ResponseWriter) {
	t := "StartTime: " + c.Time.Format("2006-01-02 15:04:05") + "\n"
	t += "Duration: " + c.Duration.String() + "\n\n"

	t += "URL: " + c.Url + "\n\n"
	t += "Method: " + c.Method + "\n"
	if len(c.RangeInfo) > 0 {
		t += "Request Range: " + c.RangeInfo + "\n"
	}

	t += "RequestHeaders: {{{\n"
	for k, v := range c.RequestHeader {
		for _, v := range v {
			t += k + ": " + v + "\n"
		}
	}

	t += "}}}\n"

	if c.Method == "POST" && c.PostBody != nil {
		t += "POST DATA: " + text(c.PostBody) + "\n"
	}

	if len(c.ContentSource) > 0 {
		t += "\nResource: " + c.ContentSource + "\n"
	}

	t += "\n"

	t += "ResponseCode: " + strconv.Itoa(c.ResponseCode) + "\n"
	t += "received bytes: " + strconv.Itoa(len(c.Bytes)) + "\n"
	if c.Error != nil {
		t += "err: " + fmt.Sprintf("%v", c.Error) + "\n"
	}

	t += "\nResponseHeaders: {{{\n"
	for k, v := range c.ResponseHeader {
		for _, v := range v {
			t += k + ": " + v + "\n"
		}
	}

	t += "}}}\n"

	if c.Bytes != nil {
		t += "Resp Data: {{{{{\n" + text(c.Bytes) + "\n}}}}}\n"
	} else {
		t += "Resp Data none\n"
	}

	fmt.Fprintln(w, t)
}

func isASCII(t []byte) bool {
	for _, b := range t {
		if b <= 6 { // ascii 0~6
			return false
		}
	}

	return true
}

func text(t []byte) string {
	suffix := ""
	if len(t) > 1024*100 {
		t = t[:1024*100]
		suffix = "..."
	}

	if isASCII(t) {
		return string(t) + suffix
	} else {
		return url.QueryEscape(string(t)) + suffix
	}
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

func (c *Cache) Save(cache *UrlCache, save bool) uint32 {
	if save {
		c.contents[cache.Url] = *cache
	}

	id := c.historyID()
	ids, ok := c.urlIds[cache.Url]
	if !ok {
		ids = make([]uint32, 0)
	}

	c.urlIds[cache.Url] = append(ids, id)
	c.indexes = append(c.indexes, &UrlHistory{*cache, id})
	return id
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
