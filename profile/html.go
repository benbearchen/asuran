package profile

import (
	"fmt"
	"html/template"
	"io"
)

type urlActionData struct {
	Pattern string
	Action  string
	Delay   string
	Edit    string
	Delete  string
	Even    bool
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
	Name    string
	Ip      string
	Owner   string
	Path    string
	Urls    []urlActionData
	Domains []domainData
}

func (p *Profile) formatViewData() profileData {
	name := p.Name
	ip := p.Ip
	owner := p.Owner
	path := p.Ip
	urls := make([]urlActionData, 0, len(p.Urls))
	domains := make([]domainData, 0, len(p.Domains))

	even := true
	for _, u := range p.Urls {
		even = !even
		urls = append(urls, urlActionData{u.UrlPattern, u.Act.String(), u.Delay.String(), u.EditCommand(), u.DeleteCommand(), even})
	}

	even = true
	for _, d := range p.Domains {
		even = !even
		domains = append(domains, domainData{d.Domain, d.Act.String(), d.TargetString(), d.EditCommand(), d.DeleteCommand(), even})
	}

	return profileData{name, ip, owner, path, urls, domains}
}

func (p *Profile) WriteHtml(w io.Writer) {
	t, err := template.ParseFiles("template/profile.tmpl")
	err = t.Execute(w, p.formatViewData())
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
