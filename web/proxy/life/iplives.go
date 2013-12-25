package life

type IPLives struct {
	lives map[string]*Life
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

	f, ok := v.lives[ip]
	if !ok {
		return nil
	} else {
		return f
	}
}
