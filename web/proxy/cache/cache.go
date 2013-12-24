package cache

type Cache struct {
	contents map[string]urlContent
	c        chan interface{}
}

type urlContent struct {
	Url   string
	Bytes []byte
}

type take struct {
	Url  string
	Take chan urlContent
}

type clear struct {
}

func NewCache() *Cache {
	c := new(Cache)
	c.contents = make(map[string]urlContent)
	c.c = make(chan interface{})

	go cacheWorker(c)
	return c
}

func (c *Cache) Save(url string, bytes []byte) {
	c.c <- urlContent{url, bytes}
}

func (c *Cache) Take(url string) ([]byte, bool) {
	urlChan := make(chan urlContent)
	c.c <- take{url, urlChan}
	content := <-urlChan
	if content.Url == url {
		return content.Bytes, true
	}

	return []byte{}, false
}

func (c *Cache) Clear() {
	c.c <- clear{}
}

func cacheWorker(c *Cache) {
	for {
		select {
		case e, ok := <-c.c:
			if !ok {
				return
			}

			switch e := e.(type) {
			case urlContent:
				c.contents[e.Url] = e
			case take:
				if content, ok := c.contents[e.Url]; ok {
					e.Take <- content
				} else {
					e.Take <- urlContent{}
				}
			case clear:
				c.contents = make(map[string]urlContent)
			}
		}
	}
}
