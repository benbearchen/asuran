package profile

type urlOperator struct {
	p *IpProfiles
}

func (u *urlOperator) Action(ip, url string) UrlProxyAction {
	profile := u.p.FindByIp(ip)
	if profile != nil {
		return profile.UrlAction(url)
	} else {
		return MakeEmptyUrlProxyAction()
	}
}

func (u *urlOperator) Delay(ip, url string) DelayAction {
	profile := u.p.FindByIp(ip)
	if profile != nil {
		return profile.UrlDelay(url)
	} else {
		return MakeEmptyDelay()
	}
}

func (u *urlOperator) Speed(ip, url string) SpeedAction {
	profile := u.p.FindByIp(ip)
	if profile != nil {
		return profile.UrlSpeed(url)
	} else {
		return MakeEmptySpeed()
	}
}

func (p *IpProfiles) OperatorUrl() UrlOperator {
	o := urlOperator{p}
	return &o
}

type profileOperator struct {
	p *IpProfiles
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

func (p *profileOperator) Owner(owner string) []*Profile {
	return p.p.FindByOwner(owner)
}

func (p *profileOperator) All() []*Profile {
	return p.p.All()
}

func (p *IpProfiles) OperatorProfile() ProfileOperator {
	o := profileOperator{p}
	return &o
}

type domainOperator struct {
	p *IpProfiles
}

func (d *domainOperator) Action(ip, domain string) *DomainAction {
	profile := d.p.FindByIp(ip)
	if profile != nil {
		return profile.Domain(domain)
	} else {
		return nil
	}
}

func (p *IpProfiles) OperatorDomain() DomainOperator {
	o := domainOperator{p}
	return &o
}
