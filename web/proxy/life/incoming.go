package life

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	uid uint64 = 1
)

func uniqueID() string {
	id := atomic.AddUint64(&uid, 1)
	return strconv.FormatUint(id, 10)
}

type Incoming struct {
	uniqueID string
	start    time.Time
	end      time.Time
	ender    func(id string)

	key string
	act string
}

func newIncoming(key, act string, ender func(id string)) *Incoming {
	c := new(Incoming)
	c.uniqueID = uniqueID()
	c.start = uniqueTime()
	c.ender = ender
	c.key = key
	c.act = act

	return c
}

func (c *Incoming) after(t time.Time) bool {
	if t.IsZero() {
		return true
	} else if !c.end.IsZero() {
		return c.end.After(t)
	} else {
		return c.start.After(t)
	}
}

func (c *Incoming) Done() {
	c.ender(c.uniqueID)
}

func (c *Incoming) ID() string {
	return c.uniqueID
}

func (c *Incoming) Start() time.Time {
	return c.start
}

func (c *Incoming) End() time.Time {
	return c.end
}

func (c *Incoming) T() time.Time {
	if !c.end.IsZero() {
		return c.end
	} else {
		return c.start
	}
}

func (c *Incoming) Key() string {
	return c.key
}

func (c *Incoming) Act() string {
	return c.act
}

type incomings struct {
	mutex sync.Mutex
	cs    map[string]*Incoming
	e     chan<- interface{}
}

func newIncomings(e chan<- interface{}) *incomings {
	c := new(incomings)
	c.cs = make(map[string]*Incoming)
	c.e = e

	return c
}

func (c *incomings) create(key, act string) *Incoming {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	in := newIncoming(key, act, c.ender)
	c.cs[in.uniqueID] = in
	return in
}

func (c *incomings) ender(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	in, ok := c.cs[id]
	if ok {
		in.end = uniqueTime()
		go func() {
			c.e <- in
		}()
	}
}

func (c *incomings) after(t time.Time) []*Incoming {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ins := make([]*Incoming, 0)
	for _, in := range c.cs {
		if in.after(t) {
			ins = append(ins, in)
		}
	}

	return ins
}
