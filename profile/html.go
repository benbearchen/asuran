package profile

import (
	"fmt"
	"html/template"
	"io"
)

type urlActionData struct {
	Pattern  string
	Action   string
	Delay    string
	Settings string
	Edit     string
	Delete   string
	Even     bool
}

type domainData struct {
	Domain string
	Action string
	IP     string
	Edit   string
	Delete string
	Even   bool
}

type profileData struct {
	Name      string
	IP        string
	Owner     string
	NotOwner  bool
	Operators string
	Path      string
	Urls      []urlActionData
	Domains   []domainData
	Stores    []string
}

func (p *Profile) formatViewData(savedIDs []string, canOperate bool) profileData {
	name := p.Name
	ip := p.Ip
	owner := p.Owner
	notOwner := !canOperate
	path := p.Ip
	urls := make([]urlActionData, 0, len(p.Urls))
	domains := make([]domainData, 0, len(p.Domains))

	operators := ""
	for op, _ := range p.Operators {
		if len(operators) == 0 {
			operators = op
		} else {
			operators += ", " + op
		}
	}

	even := true
	for _, u := range p.Urls {
		even = !even
		extra := u.Speed.String()
		if len(extra) > 0 {
			extra = ", " + extra
		}

		urls = append(urls, urlActionData{u.UrlPattern, u.Act.String(), u.Delay.String() + extra, u.Settings.String(), u.EditCommand(), u.DeleteCommand(), even})
	}

	even = true
	for _, d := range p.Domains {
		even = !even
		domains = append(domains, domainData{d.Domain, d.Act.String(), d.TargetString(), d.EditCommand(), d.DeleteCommand(), even})
	}

	return profileData{name, ip, owner, notOwner, operators, path, urls, domains, savedIDs}
}

func (p *Profile) WriteHtml(w io.Writer, savedIDs []string, realOwner bool) {
	t, err := template.ParseFiles("template/profile.tmpl")
	err = t.Execute(w, p.formatViewData(savedIDs, realOwner))
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

func WriteCommandUsage(w io.Writer) {
	t, err := template.New("profile-command-usage").Parse(`<html><body><pre>{{.}}</pre></body><html>`)
	err = t.Execute(w, CommandUsage())
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

type ownerIPData struct {
	Even bool
	Name string
	IP   string
}

type ownerData struct {
	OwnerIP string
	IPs     []ownerIPData
}

func formatOwnerData(ownerIP string, profiles []*Profile) ownerData {
	ips := make([]ownerIPData, 0)
	if profiles != nil && len(profiles) > 0 {
		even := true
		for _, p := range profiles {
			if p.Ip == "localhost" {
				continue
			}

			even = !even
			ips = append(ips, ownerIPData{even, p.Name, p.Ip})
		}
	}

	return ownerData{ownerIP, ips}
}

func WriteOwnerHtml(w io.Writer, ownerIP string, profiles []*Profile) {
	t, err := template.ParseFiles("template/owner.tmpl")
	err = t.Execute(w, formatOwnerData(ownerIP, profiles))
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

type profileDNSData struct {
	Name    string
	Host    string
	Domains []domainData
}

func formatProfileDNSData(p *Profile, host string) profileDNSData {
	domains := make([]domainData, 0, len(p.Domains))
	even := true
	for _, d := range p.Domains {
		even = !even
		domains = append(domains, domainData{d.Domain, d.Act.String(), d.TargetString(), d.EditCommand(), d.DeleteCommand(), even})
	}

	return profileDNSData{p.Name, host, domains}
}

func (p *Profile) WriteDNS(w io.Writer, host string) {
	t, err := template.ParseFiles("template/dns.tmpl")
	err = t.Execute(w, formatProfileDNSData(p, host))
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}
