package proxy

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type htmlUsage struct {
	IP string
}

func (p *Proxy) WriteUsage(w io.Writer) {
	t, err := template.ParseFiles("template/usage.tmpl")

	u := htmlUsage{}
	u.IP = p.serveIP

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

type indexData struct {
	Version string
	ServeIP string
}

func (p *Proxy) index(w http.ResponseWriter) {
	t, err := template.ParseFiles("template/index.tmpl")
	err = t.Execute(w, indexData{"0.1", p.serveIP})
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

func (p *Proxy) res(w http.ResponseWriter, r *http.Request, path string) {
	f := filepath.Clean(path)
	h := filepath.Clean("/res/")
	if !strings.HasPrefix(f, h) || len(f) <= len(h) {
		w.WriteHeader(403)
		return
	}

	dir, err := os.Getwd()
	if err != nil {
		w.WriteHeader(404)
		return
	}

	f = filepath.Join(dir, "template", f)
	http.ServeFile(w, r, f)
}
