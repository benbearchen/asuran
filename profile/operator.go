package profile

type urlOperator struct {
	p *IpProfiles
}

func (u *urlOperator) Action(ip, url string) UrlProxyAction {
	profile, ok := u.p.profiles[ip]
	if ok {
		return profile.UrlAction(url)
	} else {
		return MakeEmptyUrlProxyAction()
	}
}

func (u *urlOperator) Delay(ip, url string) DelayAction {
	profile, ok := u.p.profiles[ip]
	if ok {
		return profile.UrlDelay(url)
	} else {
		return MakeEmptyDelay()
	}
}

func (p *IpProfiles) OperatorUrl() UrlOperator {
	o := urlOperator{p}
	return &o
}

type profileOperator struct {
	p *IpProfiles
}

func (p *profileOperator) Json(ip string) string {
	profile, ok := p.p.profiles[ip]
	if ok {
		return profile.EncodeJson()
	} else {
		return ""
	}
}

func (p *profileOperator) FindByIp(ip string) *Profile {
	return p.p.FindByIp(ip)
}

func (p *profileOperator) FindByName(name string) *Profile {
	return p.p.FindByName(name)
}

func (p *profileOperator) FindByOwner(owner string) *Profile {
	s := p.p.FindByOwner(owner)
	if len(s) > 0 {
		return s[0]
	} else {
		return nil
	}
}

func (p *profileOperator) Open(ip string) *Profile {
	exists := p.FindByIp(ip)
	if exists != nil {
		return exists
	}

	return p.p.CreateProfile("", ip, "")
}

func (p *IpProfiles) OperatorProfile() ProfileOperator {
	o := profileOperator{p}
	return &o
}

type domainOperator struct {
	p *IpProfiles
}

func (p *domainOperator) Action(ip, domain string) DomainAction {
	profile, ok := p.p.profiles[ip]
	if ok {
		return profile.Domain(domain)
	} else {
		return DomainAction{}
	}
}

func (p *IpProfiles) OperatorDomain() DomainOperator {
	o := domainOperator{p}
	return &o
}
