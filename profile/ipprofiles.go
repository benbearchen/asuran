package profile

import (
	"sync"
)

type ProfilesConfig struct {
	DefaultCopyProfile string
}

type IpProfiles struct {
	profiles map[string]*Profile
	proxyOp  ProxyHostOperator
	config   ProfilesConfig

	lock sync.RWMutex
}

type ProfileOperator interface {
	FindByIp(ip string) *Profile
	FindByName(name string) *Profile
	FindByOwner(ip string) *Profile
	Open(ip string) *Profile
	Owner(owner string) []*Profile
	All() []*Profile
}

func NewIpProfiles() *IpProfiles {
	p := new(IpProfiles)
	p.profiles = make(map[string]*Profile)
	return p
}

func (p *IpProfiles) BindProxyHostOperator(op ProxyHostOperator) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.proxyOp = op
}

func (p *IpProfiles) SetDefaultCopyProfile(defaultProfile string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.config.DefaultCopyProfile = defaultProfile
}

func (p *IpProfiles) CreateProfile(name, ip, owner string) *Profile {
	p.lock.Lock()
	defer p.lock.Unlock()

	profile, ok := p.profiles[ip]
	if ok {
		// TODO: update name, owner ?
		return profile
	}

	defProf := p.config.DefaultCopyProfile
	if len(defProf) > 0 && ip != defProf {
		def, ok := p.profiles[defProf]
		if ok {
			profile = def.CloneNew(name, ip, owner)
		}
	}

	if profile == nil {
		profile = NewProfile(name, ip, owner)
	}

	profile.proxyOp = p.proxyOp
	p.profiles[ip] = profile

	return profile
}

func (p *IpProfiles) Delete(ip string) bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	_, ok := p.profiles[ip]
	if ok {
		delete(p.profiles, ip)
	}

	return ok
}

func (p *IpProfiles) FindByIp(ip string) *Profile {
	p.lock.RLock()
	defer p.lock.RUnlock()

	profile, ok := p.profiles[ip]
	if ok {
		return profile
	} else {
		return nil
	}
}

func (p *IpProfiles) FindByName(name string) *Profile {
	p.lock.RLock()
	defer p.lock.RUnlock()

	for _, profile := range p.profiles {
		if profile.Name == name {
			return profile
		}
	}

	return nil
}

func (p *IpProfiles) FindByOwner(owner string) []*Profile {
	p.lock.RLock()
	defer p.lock.RUnlock()

	profiles := []*Profile{}
	for _, profile := range p.profiles {
		if profile.Owner == owner {
			profiles = append(profiles, profile)
		}
	}

	return profiles
}

func (p *IpProfiles) CloneByName(name, newName, newIp string) *Profile {
	n := p.FindByName(name)
	if n != nil {
		p.lock.Lock()
		defer p.lock.Unlock()

		n = n.CloneNew(newName, newIp, n.Owner)
		p.profiles[newIp] = n
	}

	return n
}

func (p *IpProfiles) All() []*Profile {
	p.lock.RLock()
	defer p.lock.RUnlock()

	profiles := make([]*Profile, 0, len(p.profiles))
	for _, profile := range p.profiles {
		profiles = append(profiles, profile)
	}

	return profiles
}

func (p *IpProfiles) Load(path string) {
}

func (p *IpProfiles) Save(path string) {
}
