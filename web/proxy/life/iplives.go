package life

import (
	"sync"
	"time"
)

type IPLives struct {
	lives map[string]*Life

	lock sync.RWMutex
}

func NewIPLives() *IPLives {
	lives := IPLives{}
	lives.lives = make(map[string]*Life)
	go lives.run()
	return &lives
}

func (v *IPLives) Open(ip string) *Life {
	if len(ip) == 0 {
		return nil
	}

	if f := v.OpenExists(ip); f != nil {
		return f
	}

	v.lock.Lock()
	defer v.lock.Unlock()
	f, ok := v.lives[ip]
	if !ok {
		f = NewLife(ip)
		v.lives[ip] = f
	}

	return f
}

func (v *IPLives) Visit(ip string) *Life {
	f := v.OpenExists(ip)
	if f != nil {
		f.visit()
	}

	return f
}

func (v *IPLives) OpenExists(ip string) *Life {
	if len(ip) == 0 {
		return nil
	}

	v.lock.RLock()
	defer v.lock.RUnlock()
	f, ok := v.lives[ip]
	if !ok {
		return nil
	} else {
		return f
	}
}

func (v *IPLives) run() {
	for {
		select {
		case <-time.NewTimer(time.Second * 15).C:
			v.checkIdle()
		}
	}
}

func (v *IPLives) checkIdle() {
	v.lock.RLock()
	defer v.lock.RUnlock()
	for _, f := range v.lives {
		if f.isIdle(time.Minute * 30) {
			f.Restart()
		}
	}
}
