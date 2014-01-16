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
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Proxy struct {
	webServers map[int]*httpd.Http
	lives      *life.IPLives
	urlOp      profile.UrlOperator
	profileOp  profile.ProfileOperator
	domainOp   profile.DomainOperator
	serveIP    string
	mainHost   string

	lock sync.RWMutex
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
			fmt.Println("proxy on ip: " + ip)
			for port, _ := range p.webServers {
				p.mainHost = ip
				if port != 80 {
					p.mainHost += ":" + strconv.Itoa(port)
				}

				fmt.Println("visit http://" + p.mainHost + "/ for more information")
			}
		}
	}

	return p
}

func (p *Proxy) Bind(port int) {
	p.lock.Lock()
	defer p.lock.Unlock()

	h := &httpd.Http{}
	p.webServers[port] = h
	serverAddress := fmt.Sprintf(":%d", port)
	h.Init(serverAddress)
	h.RegisterHandler(p)
	go h.Run()
}

func (p *Proxy) TryBind(port int) {
	exists := false
	func() {
		p.lock.RLock()
		defer p.lock.RUnlock()

		_, exists = p.webServers[port]
	}()

	if !exists {
		p.Bind(port)
	}
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
	targetHost := httpd.RemoteHost(r.Host)
	remoteIP := httpd.RemoteHost(r.RemoteAddr)
	urlPath := r.URL.Path
	//fmt.Printf("host: %s/%s, remote: %s/%s, url: %s\n", targetHost, r.Host, remoteIP, r.RemoteAddr, urlPath)
	if strings.HasPrefix(urlPath, "http://") {
		p.proxyUrl(urlPath, w, r)
	} else if targetHost == "i.me" {
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
	} else if _, m := httpd.MatchPath(urlPath, "/res"); m {
		p.res(w, r, urlPath)
	} else if urlPath == "/" {
		p.index(w)
	} else if urlPath == "/devices" {
		p.devices(w)
	} else if urlPath == "/about" {
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
		fmt.Fprintln(w, "visit http://"+p.mainHost+"/about to get info")
		fmt.Fprintln(w, "visit http://"+p.mainHost+"/test/"+p.mainHost+"/about to test the proxy of http://"+p.mainHost+"/about")
		fmt.Fprintln(w, "visit http://"+p.mainHost+"/to/"+p.mainHost+"/about to purely proxy of http://"+p.mainHost+"/about")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "visit http://"+p.mainHost+"/test/"+p.mainHost+"/test/"+p.mainHost+"/about to test the proxy")
	} else if targetHost == "localhost" || targetHost == "127.0.0.1" || targetHost == p.serveIP {
		fmt.Fprintln(w, "visit http://"+r.Host+"/about to get info")
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

	p.profileOp.Open(remoteIP)

	rangeInfo := cache.CheckRange(r)
	f := p.lives.Open(remoteIP)
	var u *life.UrlState
	if f != nil {
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

	if needCache && r.Method == "GET" && f != nil {
		c := f.CheckCache(fullUrl, rangeInfo)
		if c != nil && c.Error == nil {
			c.Response(w)
			return
		}
	}

	httpStart := time.Now()
	resp, postBody, err := net.NewHttp(fullUrl, r, p.parseDomainAsDial(target, remoteIP))
	if err != nil {
		c := cache.NewUrlCache(fullUrl, r, nil, nil, nil, rangeInfo, httpStart, time.Now(), err)
		if f != nil {
			go p.saveContentToCache(fullUrl, f, c, false)
		}

		http.Error(w, "Bad Gateway", 502)
	} else {
		defer resp.Close()
		content, err := resp.ProxyReturn(w)
		httpEnd := time.Now()
		c := cache.NewUrlCache(fullUrl, r, postBody, resp, content, rangeInfo, httpStart, httpEnd, err)
		if f != nil {
			go p.saveContentToCache(fullUrl, f, c, needCache)
		}
	}
}

func (p *Proxy) initDevice(w io.Writer, ip string) {
	if p.profileOp != nil {
		p.profileOp.Open(ip)
		p.lives.Open(ip)
		p.WriteInitDevice(w, ip)
	} else {
		fmt.Fprintln(w, "sorry, can't create profile for you!")
	}
}

type proxyDomainOperator struct {
	p *Proxy
}

func (p *proxyDomainOperator) Action(ip, domain string) *profile.DomainAction {
	if domain == "i.me" {
		p.p.LogDomain(ip, "init", domain, p.p.serveIP)
		return profile.NewDomainAction(domain, profile.DomainActRedirect, p.p.serveIP)
	} else if p.p.domainOp != nil {
		act := "query"
		a := p.p.domainOp.Action(ip, domain)
		if a != nil {
			b := *a
			a = &b
		}

		if a != nil && a.Act == profile.DomainActRedirect && a.IP == "" {
			a.IP = p.p.serveIP
			act = "proxy"
		}

		resultIP := ""
		if a != nil && len(a.IP) > 0 {
			resultIP = a.IP
		}

		p.p.LogDomain(ip, act, domain, resultIP)
		return a
	} else {
		p.p.LogDomain(ip, "undef", domain, "")
		return nil
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

	if op == "" && f.Owner == "" && ownerIP != profileIP {
		f.Owner = ownerIP
		p.lives.Open(profileIP)
	}

	if op == "export" {
		fmt.Fprintln(w, f.ExportCommand())
		return
	} else if op == "restart" {
		if f := p.lives.OpenExists(profileIP); f != nil {
			f.Restart()
			fmt.Fprintln(w, profileIP+" 已经重新初始化")
		} else {
			fmt.Fprintln(w, profileIP+" 不存在")
		}

		fmt.Fprintln(w, "# 关了这个窗口吧 #")
		return
	} else if op == "history" {
		if f := p.lives.OpenExists(profileIP); f != nil {
			p.writeHistory(w, profileIP, f)
		} else {
			fmt.Fprintln(w, profileIP+" 不存在")
		}
		return
	} else if op == "look" || op == "list" || op == "detail" {
		if len(pages) >= 4 {
			id, err := strconv.ParseInt(pages[3], 10, 32)
			if err == nil {
				p.lookHistoryByID(w, profileIP, uint32(id), op)
				return
			}
		}

		lookUrl := ""
		if len(pages) >= 4 {
			lookUrl = "http://" + strings.Join(pages[3:], "/")
		}

		if len(r.URL.RawQuery) > 0 {
			lookUrl += "?" + r.URL.RawQuery
		}

		p.lookHistory(w, profileIP, lookUrl, op)
		return
	}

	r.ParseForm()
	if v, ok := r.Form["cmd"]; ok && len(v) > 0 {
		if f.Owner == "" || f.Owner == ownerIP {
			for _, cmd := range v {
				p.Command(cmd, f, p.lives.Open(profileIP))
			}
		}
	}

	f.WriteHtml(w)
}

func (p *Proxy) lookHistoryByID(w http.ResponseWriter, profileIP string, id uint32, op string) {
	f := p.lives.OpenExists(profileIP)
	if f == nil {
		fmt.Fprintln(w, profileIP+" 不存在")
		return
	}

	h := f.LookHistoryByID(id)
	if h != nil {
		switch op {
		case "look":
			h.Response(w)
		case "detail":
			h.Detail(w)
		}
	} else {
		fmt.Fprintf(w, "history %d not exist", id)
	}
}

func (p *Proxy) lookHistory(w http.ResponseWriter, profileIP, lookUrl, op string) {
	f := p.lives.OpenExists(profileIP)
	if f == nil {
		fmt.Fprintln(w, profileIP+" 不存在")
		return
	}

	switch op {
	case "look":
		if c := f.LookCache(lookUrl); c != nil {
			c.Response(w)
		} else {
			fmt.Fprintln(w, "can't look up "+lookUrl)
		}
	case "detail":
		if c := f.LookCache(lookUrl); c != nil {
			c.Detail(w)
		} else {
			fmt.Fprintln(w, "can't look up "+lookUrl+" for detail")
		}
	case "list":
		histories := f.ListHistory(lookUrl)
		p.writeUrlHistoryList(w, profileIP, lookUrl, histories)
	}
}

func (p *Proxy) LogDomain(ip, action, domain, resultIP string) {
	if f := p.lives.Open(ip); f != nil {
		s := ""
		if len(resultIP) > 0 {
			s = " "
		}

		f.Log("domain " + action + " " + domain + s + resultIP)
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

func (p *Proxy) saveContentToCache(fullUrl string, f *life.Life, c *cache.UrlCache, needCache bool) {
	id := f.SaveContentToCache(c, needCache)

	info := "proxy " + fullUrl + " " + strconv.FormatUint(uint64(id), 10)
	if len(c.RangeInfo) > 0 {
		info += " " + c.RangeInfo
	}

	f.Log(info)
}

type proxyHostOperator struct {
	p *Proxy
}

func (p *proxyHostOperator) New(port int) {
	p.p.TryBind(port)
}

func (p *Proxy) NewProxyHostOperator() profile.ProxyHostOperator {
	return &proxyHostOperator{p}
}

func (p *Proxy) parseDomainAsDial(target, client string) func(network, addr string) (gonet.Conn, error) {
	if p.domainOp == nil {
		return nil
	}

	u, err := url.Parse(target)
	if err != nil {
		return nil
	}

	domain, port, err := gonet.SplitHostPort(u.Host)
	if err != nil {
		domain = u.Host
	}

	if len(port) == 0 {
		port = "80"
	}

	a := p.domainOp.Action(client, domain)
	if a == nil || a.Act != profile.DomainActRedirect || len(a.IP) == 0 || a.IP == p.serveIP {
		return nil
	}

	address := gonet.JoinHostPort(a.IP, port)
	return func(network, addr string) (gonet.Conn, error) {
		if network == "tcp" {
			return gonet.Dial(network, address)
		} else {
			return gonet.Dial(network, addr)
		}
	}
}
