package proxy

import (
	"fmt"
	"html/template"
	"io"
)

type htmlUsage struct {
	IpInfo string
}

func (p *Proxy) WriteUsage(w io.Writer) {
	t, err := template.ParseFiles("template/usage.tmpl")

	u := htmlUsage{}
	u.IpInfo = "（" + p.serveIP + "）"

	err = t.Execute(w, u)
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

type htmlInitDevice struct {
	ProxyIP  string
	ClientIP string
}

func (p *Proxy) WriteInitDevice(w io.Writer, ip string) {
	t, err := template.ParseFiles("template/i.me.tmpl")

	d := htmlInitDevice{}
	d.ProxyIP = p.serveIP
	d.ClientIP = ip

	err = t.Execute(w, d)
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}
