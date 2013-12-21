package profile

import (
	"encoding/json"
)

type IpProfiles struct {
	profiles map[string]*Profile
}

type ProfileOperator interface {
	Json(ip string) string
	FindByIp(ip string) *Profile
	FindByName(name string) *Profile
	FindByOwner(ip string) *Profile
	Open(ip string) *Profile
	Owner(owner string) []*Profile
}

func NewIpProfiles() *IpProfiles {
	p := new(IpProfiles)
	p.profiles = make(map[string]*Profile)
	return p
}

func (p *IpProfiles) CreateProfile(name, ip, owner string) *Profile {
	profile := NewProfile(name, ip, owner)
	p.profiles[ip] = profile
	return profile
}

func (p *IpProfiles) FindByIp(ip string) *Profile {
	profile, ok := p.profiles[ip]
	if ok {
		return profile
	} else {
		return nil
	}
}

func (p *IpProfiles) FindByName(name string) *Profile {
	for _, profile := range p.profiles {
		if profile.Name == name {
			return profile
		}
	}

	return nil
}

func (p *IpProfiles) FindByOwner(owner string) []*Profile {
	profiles := []*Profile{}
	for _, profile := range p.profiles {
		if profile.Owner == owner {
			profiles = append(profiles, profile)
		}
	}

	return profiles
}

func (p *IpProfiles) CloneByName(name, newName, newIp string) *Profile {
	var n *Profile = nil
	for _, profile := range p.profiles {
		if profile.Name == name {
			n = profile.CloneNew(newName, newIp)
			break
		}
	}

	if n != nil {
		p.profiles[newIp] = n
	}

	return n
}

func (p *IpProfiles) EncodeJson() string {
	bytes, err := json.Marshal(p.profiles)
	if err != nil {
		return ""
	} else {
		return string(bytes)
	}
}

func (p *IpProfiles) DecodeJson(f string) {
}

func (p *IpProfiles) Load(path string) {
}

func (p *IpProfiles) Save(path string) {
}
