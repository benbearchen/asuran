package cache

type UrlCache struct {
	Url   string
	Bytes []byte
}

type Cache struct {
	contents map[string]UrlCache
}

func NewCache() *Cache {
	c := new(Cache)
	c.contents = make(map[string]UrlCache)

	return c
}

func (c *Cache) Save(url string, bytes []byte) {
	c.contents[url] = UrlCache{url, bytes}
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
