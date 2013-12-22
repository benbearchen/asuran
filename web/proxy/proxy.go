package proxy

import (
	"github.com/benbearchen/asuran/net"
	"github.com/benbearchen/asuran/net/httpd"
	"github.com/benbearchen/asuran/profile"
	"github.com/benbearchen/asuran/web/proxy/cache"

	"fmt"
	"io"
	gonet "net"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type Proxy struct {
	webServers map[int]*httpd.Http
	cache      *cache.Cache
	urlOp      profile.UrlOperator
	profileOp  profile.ProfileOperator
	domainOp   profile.DomainOperator
	serveIP    string
}

func NewProxy() *Proxy {
	p := new(Proxy)
	p.webServers = make(map[int]*httpd.Http)
	p.cache = cache.NewCache()
	p.Bind(80)

	ips := net.LocalIPs()
	if ips == nil {
		fmt.Println("Proxy can't get local ip")
	} else {
		for _, ip := range ips {
			p.serveIP = ip
			fmt.Println("proxy on ip: ", ip)
		}
	}

	return p
}

func (p *Proxy) Bind(port int) {
	h := &httpd.Http{}
	p.webServers[port] = h
	serverAddress := fmt.Sprintf(":%d", port)
	h.Init(serverAddress)
	h.RegisterHandler(p)
	go h.Run()
}

func (p *Proxy) BindUrlOperator(op profile.UrlOperator) {
	p.urlOp = op
}

func (p *Proxy) BindProfileOperator(op profile.ProfileOperator) {
	p.profileOp = op
}

func (p *Proxy) BindDomainOperator(op profile.DomainOperator) {
	p.domainOp = op
}

func (p *Proxy) GetDescription() string {
	return "web transparent proxy"
}

func (p *Proxy) GetHandlePath() string {
	return "/"
}

func (p *Proxy) testUrl(
	target string,
	w http.ResponseWriter,
	r *http.Request) {
	fmt.Fprintln(w, "proxy test: "+target)

	start := time.Now()
	resp, err := net.NewHttpGet(target)
	if err != nil {
		fmt.Fprintln(w, "error: "+err.Error())
	} else {
		defer resp.Close()
		content, err := resp.ReadAll()
		end := time.Now()
		if err != nil {
			fmt.Fprintf(w, "error: %s", err.Error())
		} else {
			fmt.Fprintf(w, "goroutins count: %d\n", runtime.NumGoroutine())
			fmt.Fprintln(w, "used time: "+end.Sub(start).String())
			fmt.Fprintln(w, "content: [[[[[[[")
			fmt.Fprintln(w, content)
			fmt.Fprintln(w, "]]]]]]]")
		}
	}
}

func (p *Proxy) OnRequest(
	w http.ResponseWriter,
	r *http.Request) {
	targetHost := r.Host
	remoteIP := httpd.RemoteHost(r.RemoteAddr)
	urlPath := r.URL.Path
	if targetHost == "i.me" {
		p.initDevice(w, remoteIP)
	} else if targetHost != "localhost" && targetHost != "127.0.0.1" && targetHost != p.serveIP {
		target := "http://" + r.Host + urlPath
		p.proxyUrl(target, w, r)
	} else if page, m := httpd.MatchPath(urlPath, "/test/"); m {
		if page == "" {
			page = "localhost"
		}

		target := "http://" + page
		p.testUrl(target, w, r)
	} else if page, m := httpd.MatchPath(urlPath, "/to/"); m {
		target := "http://" + page
		p.proxyUrl(target, w, r)
	} else if _, m := httpd.MatchPath(urlPath, "/usage"); m {
		p.WriteUsage(w)
	} else if page, m := httpd.MatchPath(urlPath, "/profile"); m {
		p.ownProfile(remoteIP, page, w, r)
	} else if urlPath == "/" || urlPath == "/about" {
		scheme := r.URL.Scheme
		if scheme == "" {
			scheme = "http"
		}

		fmt.Fprintln(w, "url: "+r.Method+" "+scheme+"://"+r.Host+urlPath)
		fmt.Fprintln(w, "remote: "+r.RemoteAddr)

		for k, v := range r.Header {
			fmt.Fprintln(w, "header: "+k+": "+strings.Join(v, "|"))
		}

		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "visit http://localhost/about to get info")
		fmt.Fprintln(w, "visit http://localhost/test/localhost/about to test the proxy of http://localhost/about")
		fmt.Fprintln(w, "visit http://localhost/to/localhost/about to purely proxy of http://localhost/about")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "visit http://localhost/test/localhost/test/localhost/about to test the proxy")
	} else if r.Host == "localhost" || r.Host == "127.0.0.1" {
		fmt.Fprintln(w, "visit http://localhost/about to get info")
	} else if remoteIP == r.Host {
		// 代理本机访问……
		http.Error(w, "Not Found", 404)
	} else {
		target := "http://" + r.Host + urlPath
		p.proxyUrl(target, w, r)
	}
}

func (p *Proxy) proxyUrl(target string, w http.ResponseWriter, r *http.Request) {
	//fmt.Println("proxy: " + target)
	remoteIP := httpd.RemoteHost(r.RemoteAddr)
	needCache := false
	if p.urlOp != nil {
		act := p.urlOp.Action(remoteIP, target)
		//fmt.Println("url act: " + act.String())
		if act.Act == profile.UrlActCache {
			needCache = true
		}

		delay := p.urlOp.Delay(remoteIP, target)
		//fmt.Println("url delay: " + delay.String())
		switch delay.Act {
		case profile.DelayActDelayEach:
			if delay.Time > 0 {
				time.Sleep(delay.Duration())
			}
			break
		case profile.DelayActDropUntil:
			// TODO:
			break
		}
	}

	if needCache {
		bytes, ok := p.checkCache(target)
		if ok {
			fmt.Fprintf(w, "%s", string(bytes))
			return
		}
	}

	resp, err := net.NewHttpGet(target)
	if err != nil {
		http.Error(w, "Bad Gateway", 502)
	} else {
		defer resp.Close()
		content, err := resp.ReadAll()
		if err != nil {
			http.Error(w, "Bad Gateway", 502)
		} else {
			fmt.Fprintf(w, "%s", content)
			if needCache {
				go p.saveContentToCache(target, content)
			}
		}
	}
}

func (p *Proxy) checkCache(url string) ([]byte, bool) {
	return p.cache.Take(url)
}

func (p *Proxy) saveContentToCache(url string, content string) {
	p.cache.Save(url, []byte(content))
}

func (p *Proxy) initDevice(w io.Writer, ip string) {
	if p.profileOp != nil {
		p.profileOp.Open(ip)
		p.WriteInitDevice(w, ip)
	} else {
		fmt.Fprintln(w, "sorry, can't create profile for you!")
	}
}

type proxyDomainOperator struct {
	p *Proxy
}

func (p *proxyDomainOperator) Action(ip, domain string) profile.DomainAction {
	if domain == "i.me" {
		return profile.DomainAction{domain, profile.DomainActRedirect, p.p.serveIP}
	} else if p.p.domainOp != nil {
		a := p.p.domainOp.Action(ip, domain)
		if a.Act == profile.DomainActRedirect && a.IP == "" {
			a.IP = p.p.serveIP
		}

		return a
	} else {
		return profile.DomainAction{}
	}
}

func (p *Proxy) NewDomainOperator() profile.DomainOperator {
	o := proxyDomainOperator{p}
	return &o
}

func (p *Proxy) ownProfile(ownerIP, page string, w http.ResponseWriter, r *http.Request) {
	if p.profileOp == nil {
		fmt.Fprintln(w, "can't locate profile")
		return
	}

	if page == "/commands" {
		profile.WriteCommandUsage(w)
		return
	} else if page == "" || page == "/" {
		profiles := p.profileOp.Owner(ownerIP)
		profile.WriteOwnerHtml(w, ownerIP, profiles)
		return
	}

	profileIP := ""
	op := ""
	pages := strings.Split(page, "/")
	if len(pages) >= 2 {
		if ip := gonet.ParseIP(pages[1]); ip != nil {
			profileIP = ip.String()
		}

		if len(pages) >= 3 {
			op = pages[2]
		}
	}

	if profileIP == "" {
		profiles := p.profileOp.Owner(ownerIP)
		profile.WriteOwnerHtml(w, ownerIP, profiles)
		return
	}

	f := p.profileOp.FindByIp(profileIP)
	if f == nil {
		f = p.profileOp.Open(profileIP)
	}

	if f.Owner == "" {
		f.Owner = ownerIP
	}

	if op == "export" {
		fmt.Fprintln(w, f.ExportCommand())
		return
	}

	r.ParseForm()
	if v, ok := r.Form["cmd"]; ok && len(v) > 0 {
		if f.Owner == ownerIP {
			f.Command(v[0])
		}
	}

	f.WriteHtml(w)
}
