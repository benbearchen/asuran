package proxy

import (
	"github.com/benbearchen/asuran/net"
	"github.com/benbearchen/asuran/net/httpd"
	"github.com/benbearchen/asuran/profile"
	"github.com/benbearchen/asuran/web/proxy/cache"
	"github.com/benbearchen/asuran/web/proxy/life"

	"fmt"
	"io"
	"io/ioutil"
	gonet "net"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type Proxy struct {
	webServers map[int]*httpd.Http
	lives      *life.IPLives
	urlOp      profile.UrlOperator
	profileOp  profile.ProfileOperator
	domainOp   profile.DomainOperator
	serveIP    string
}

func NewProxy() *Proxy {
	p := new(Proxy)
	p.webServers = make(map[int]*httpd.Http)
	p.lives = life.NewIPLives()
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
	} else if _, m := httpd.MatchPath(urlPath, "/post"); m {
		p.postTest(w, r)
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

		fmt.Fprintf(w, "url%%v: %v\n", r.URL)
		fmt.Fprintln(w, "url.String(): "+r.URL.String())
		fmt.Fprintln(w)

		ex := ""
		if len(r.URL.RawQuery) > 0 {
			ex += "?" + r.URL.RawQuery
		}

		fmt.Fprintln(w, "url: "+r.Method+" "+scheme+"://"+r.Host+urlPath+ex)
		fmt.Fprintln(w, "remote: "+r.RemoteAddr)
		fmt.Fprintln(w, "requestURI: "+r.RequestURI)
		fmt.Fprintln(w, "host: "+r.Host)

		fmt.Fprintln(w)
		for k, v := range r.Header {
			fmt.Fprintln(w, "header: "+k+": "+strings.Join(v, "|"))
		}

		fmt.Fprintln(w)
		for _, s := range r.TransferEncoding {
			fmt.Fprintln(w, "transferEncoding: "+s)
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

	fullUrl := target
	if len(r.URL.RawQuery) > 0 {
		fullUrl += "?" + r.URL.RawQuery
	}

	rangeInfo := cache.CheckRange(r)
	f := p.lives.Open(remoteIP)
	var u *life.UrlState
	if f != nil {
		info := "proxy " + fullUrl
		if len(rangeInfo) > 0 {
			info += " " + rangeInfo
		}

		f.Log(info)
		u = f.OpenUrl(fullUrl)
	}

	if p.urlOp != nil {
		act := p.urlOp.Action(remoteIP, fullUrl)
		//fmt.Println("url act: " + act.String())
		if act.Act == profile.UrlActCache {
			needCache = true
		}

		delay := p.urlOp.Delay(remoteIP, fullUrl)
		//fmt.Println("url delay: " + delay.String())
		switch delay.Act {
		case profile.DelayActDelayEach:
			if delay.Time > 0 {
				// TODO: create request before sleep, more effective
				time.Sleep(delay.Duration())
			}
			break
		case profile.DelayActDropUntil:
			if u != nil {
				if u.DropUntil(delay.Duration()) {
					// TODO: more safe method, maybe net.http.Hijacker
					panic("")
				}
			}
			break
		}
	}

	if needCache && r.Method == "GET" {
		c := f.CheckCache(fullUrl, rangeInfo)
		if c != nil {
			c.Response(w)
			return
		}
	}

	resp, err := net.NewHttp(fullUrl, r)
	if err != nil {
		http.Error(w, "Bad Gateway", 502)
	} else {
		defer resp.Close()
		content, err := resp.ReadAllBytes()
		if err != nil {
			http.Error(w, "Bad Gateway", 502)
		} else {
			c := cache.NewUrlCache(fullUrl, resp, content, rangeInfo)
			c.Response(w)
			go f.SaveContentToCache(c)
		}
	}
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
		p.p.LogDomain(ip, "init", domain)
		return profile.DomainAction{domain, profile.DomainActRedirect, p.p.serveIP}
	} else if p.p.domainOp != nil {
		a := p.p.domainOp.Action(ip, domain)
		if a.Act == profile.DomainActRedirect && a.IP == "" {
			a.IP = p.p.serveIP
		}

		p.p.LogDomain(ip, "query", domain)
		return a
	} else {
		p.p.LogDomain(ip, "undef", domain)
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

	if f.Owner == "" && ownerIP != profileIP {
		f.Owner = ownerIP
	}

	if op == "export" {
		fmt.Fprintln(w, f.ExportCommand())
		return
	} else if op == "restart" {
		if f := p.lives.Open(profileIP); f != nil {
			f.Restart()
			fmt.Fprintln(w, profileIP+" 已经重新初始化")
		} else {
			fmt.Fprintln(w, profileIP+" 不存在")
		}

		fmt.Fprintln(w, "# 关了这个窗口吧 #")
		return
	} else if op == "history" {
		if f := p.lives.OpenExists(profileIP); f != nil {
			fmt.Fprintln(w, f.FormatHistory())
		} else {
			fmt.Fprintln(w, profileIP+" 不存在")
		}
		return
	} else if op == "look" {
		if f := p.lives.OpenExists(profileIP); f != nil {
			lookUrl := page
			if len(pages) >= 4 {
				lookUrl = "http://" + strings.Join(pages[3:], "/")
			}

			if len(r.URL.RawQuery) > 0 {
				lookUrl += "?" + r.URL.RawQuery
			}

			if c := f.LookCache(lookUrl); c != nil {
				c.Response(w)
			} else {
				fmt.Fprintln(w, "can't look up "+lookUrl)
			}
		} else {
			fmt.Fprintln(w, profileIP+" 不存在")
		}
		return
	}

	r.ParseForm()
	if v, ok := r.Form["cmd"]; ok && len(v) > 0 {
		if f.Owner == "" || f.Owner == ownerIP {
			f.Command(v[0])
		}
	}

	f.WriteHtml(w)
}

func (p *Proxy) LogDomain(ip, action, domain string) {
	if f := p.lives.Open(ip); f != nil {
		f.Log("domain " + action + " " + domain)
	}
}

func (p *Proxy) postTest(w http.ResponseWriter, r *http.Request) {
	head := "<html><body>"
	if r.Method == "POST" {
		b, err := ioutil.ReadAll(r.Body)
		if err == nil {
			head += string(b)
			fmt.Printf("recv body: %s\n", string(b))
		}
	}

	fmt.Println()

	form := `<form method="POST" action="/post"><input type="text" name="right_items" value="5 6"><input type="submit" value="post"></form></body></html>`
	fmt.Fprintln(w, head+form)
}
