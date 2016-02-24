package profile

import (
	"github.com/benbearchen/asuran/policy"

	"fmt"
	"html/template"
	"io"
	"sort"
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

	keys := make([]string, 0, len(p.Urls)+1)
	keys = append(keys, "")
	for k, _ := range p.Urls {
		keys = append(keys, k)
	}

	sort.Strings(keys[1:])
	even := true
	for _, k := range keys {
		even = !even
		target := ""
		edit := ""
		del := ""
		var up *policy.UrlPolicy
		if k == "" {
			up = p.UrlDefault
			target = "[缺省目标]"
		} else {
			up = p.Urls[k].p
			target = up.Target()
			del = "url delete " + target + "\n"
		}

		act := up.ContentComment()
		delay := up.DelayComment()
		other := up.OtherComment()

		edit = up.Command() + "\n"

		urls = append(urls, urlActionData{target, act, delay, other, edit, del, even})
	}

	keys = make([]string, 0, len(p.Domains))
	for k, _ := range p.Domains {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	even = true
	for _, k := range keys {
		d := p.Domains[k]
		even = !even
		act := d.p.Comment()
		edit := d.p.Command() + "\n"
		del := "domain delete " + d.p.Domain() + "\n"

		domains = append(domains, domainData{d.Domain, act, d.TargetString(), edit, del, even})
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
	err = t.Execute(w, policy.CommandUsage())
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
		act := d.p.Comment()
		edit := d.p.Command() + "\n"
		del := "domain delete " + d.p.Domain() + "\n"

		domains = append(domains, domainData{d.Domain, act, d.TargetString(), edit, del, even})
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
