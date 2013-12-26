package life

import (
	"sync"
)

type IPLives struct {
	lives map[string]*Life

	lock sync.RWMutex
}

func NewIPLives() *IPLives {
	lives := IPLives{}
	lives.lives = make(map[string]*Life)
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
