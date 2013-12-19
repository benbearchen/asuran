package cache

type Cache struct {
	contents map[string]urlContent
	save     chan urlContent
	take     chan take
}

type urlContent struct {
	Url   string
	Bytes []byte
}

type take struct {
	Url  string
	Take chan urlContent
}

func NewCache() *Cache {
	c := new(Cache)
	c.contents = make(map[string]urlContent)
	c.save = make(chan urlContent)
	c.take = make(chan take)

	go cacheWorker(c)
	return c
}

func (c *Cache) Save(url string, bytes []byte) {
	c.save <- urlContent{url, bytes}
}

func (c *Cache) Take(url string) ([]byte, bool) {
	urlChan := make(chan urlContent)
	c.take <- take{url, urlChan}
	content := <-urlChan
	if content.Url == url {
		return content.Bytes, true
	}

	return []byte{}, false
}

func cacheWorker(c *Cache) {
	for {
		select {
		case content, ok := <-c.save:
			if !ok {
				return
			} else {
				c.contents[content.Url] = content
			}
		case take, ok := <-c.take:
			if !ok {
				return
			} else if content, ok := c.contents[take.Url]; ok {
				take.Take <- content
			} else {
				take.Take <- urlContent{}
			}
		}
	}
}
