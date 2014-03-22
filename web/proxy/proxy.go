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
	"math/rand"
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
	r    *rand.Rand
}

func NewProxy() *Proxy {
	p := new(Proxy)
	p.webServers = make(map[int]*httpd.Http)
	p.lives = life.NewIPLives()
	p.r = rand.New(rand.NewSource(time.Now().UnixNano()))

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
	p.profileOp.Open("localhost").Name = "DNS 服务"
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
	} else if !p.isSelfAddr(targetHost) && !p.isSelfAddr(remoteIP) {
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
	} else if page, m := httpd.MatchPath(urlPath, "/dns"); m {
		p.dns(page, w, r)
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
	} else if urlPath == "/urlencoded" {
		p.urlEncoded(w)
	} else {
		fmt.Fprintln(w, "visit http://"+r.Host+"/about to get info")
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

	requestUrl := fullUrl
	requestR := r
	contentSource := ""

	var prof *profile.Profile
	if !p.isSelfAddr(remoteIP) {
		prof = p.profileOp.Open(remoteIP)
	}

	rangeInfo := cache.CheckRange(r)
	f := p.lives.Open(remoteIP)
	var u *life.UrlState
	if f != nil {
		u = f.OpenUrl(fullUrl)
	}

	if p.urlOp != nil {
		delay := p.urlOp.Delay(remoteIP, fullUrl)
		//fmt.Println("url delay: " + delay.String())
		switch delay.Act {
		case profile.DelayActDelayEach:
			if delay.Time > 0 {
				// TODO: create request before sleep, more effective
				d := delay.RandDuration(p.r)
				time.Sleep(d)
				f.Log("proxy " + fullUrl + " delay " + d.String())
			}
			break
		case profile.DelayActDropUntil:
			d := delay.RandDuration(p.r)
			if u != nil && u.DropUntil(d) {
				// TODO: more safe method, maybe net.http.Hijacker
				f.Log("proxy " + fullUrl + " drop " + d.String())
				panic("")
			}
			break
		case profile.DelayActTimeout:
			if delay.Time > 0 {
				// TODO: more safe method, maybe net.http.Hijacker
				d := delay.RandDuration(p.r)
				time.Sleep(d)
				f.Log("proxy " + fullUrl + " timeout " + d.String())
				panic("")
			}
			break
		}

		act := p.urlOp.Action(remoteIP, fullUrl)
		//fmt.Println("url act: " + act.String())
		switch act.Act {
		case profile.UrlActCache:
			needCache = true
		case profile.UrlActStatus:
			status := 502
			if c, err := strconv.Atoi(act.ContentValue); err == nil {
				status = c
			}

			w.WriteHeader(status)
			f.Log("proxy " + fullUrl + " status " + strconv.Itoa(status))
			return
		case profile.UrlActMap:
			requestUrl = act.ContentValue
			requestR = nil
			contentSource = "map " + act.ContentValue
		case profile.UrlActRedirect:
			http.Redirect(w, r, act.ContentValue, 302)
			f.Log("proxy " + fullUrl + " redirect " + act.ContentValue)
			return
		case profile.UrlActRewritten:
			fallthrough
		case profile.UrlActRestore:
			if p.rewriteUrl(fullUrl, w, r, rangeInfo, prof, f, act) {
				return
			}
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
	resp, postBody, err := net.NewHttp(requestUrl, requestR, p.parseDomainAsDial(target, remoteIP))
	if err != nil {
		c := cache.NewUrlCache(fullUrl, r, nil, nil, contentSource, nil, rangeInfo, httpStart, time.Now(), err)
		if f != nil {
			go p.saveContentToCache(fullUrl, f, c, false)
		}

		http.Error(w, "Bad Gateway", 502)
	} else {
		defer resp.Close()
		content, err := resp.ProxyReturn(w)
		httpEnd := time.Now()
		c := cache.NewUrlCache(fullUrl, r, postBody, resp, contentSource, content, rangeInfo, httpStart, httpEnd, err)
		if f != nil {
			go p.saveContentToCache(fullUrl, f, c, needCache)
		}
	}
}

func (p *Proxy) rewriteUrl(target string, w http.ResponseWriter, r *http.Request, rangeInfo string, prof *profile.Profile, f *life.Life, act profile.UrlProxyAction) bool {
	var content []byte = nil
	contentSource := ""
	switch act.Act {
	case profile.UrlActRewritten:
		u, err := url.QueryUnescape(act.ContentValue)
		if err != nil {
			return false
		}

		content = []byte(u)
		contentSource = "rewrite"
	case profile.UrlActRestore:
		content = prof.Restore(act.ContentValue)
		if content == nil {
			return false
		}
	default:
		return false
	}

	start := time.Now()
	w.Write(content)
	c := cache.NewUrlCache(target, r, nil, nil, contentSource, content, rangeInfo, start, time.Now(), nil)
	c.ResponseCode = 200
	if f != nil {
		p.saveContentToCache(target, f, c, false)
	}

	return true
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
	}

	if p.p.profileOp != nil {
		if p.p.profileOp.FindByIp(ip) == nil {
			a := p.p.domainOp.Action("localhost", domain)
			if a != nil {
				b := *a
				a = &b
			}

			return a
		}
	}

	if p.p.domainOp != nil {
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
		if p.isLoopback(pages[1]) {
			fmt.Fprintln(w, "can't profile localhost or 127.0.0.1")
			return
		}

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
			id, err := strconv.ParseUint(pages[3], 10, 32)
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
	} else if op == "domain" {
		if len(pages) >= 4 {
			switch pages[3] {
			case "redirect":
				if len(pages) >= 5 {
					domain := pages[4]
					f.SetUrl(profile.UrlToPattern(domain), nil, nil)
					fmt.Fprintf(w, "<html><head><title>代理域名 %s</title></head><body>域名 %s 已处理。<br/>返回 <a href=\"/profile/%s\">管理页面</a></body></html>", domain, domain, profileIP)
					return
				}
			}
		}

		fmt.Fprintf(w, "<html><body>无效请求。<br/>返回 <a href=\"/profile/%s\">管理页面</a></body></html>", profileIP)
		return
	} else if op == "url" {
		if len(pages) >= 4 {
			switch pages[3] {
			case "store":
				if len(pages) >= 5 {
					id := pages[4]
					if u, sid := p.storeHistory(profileIP, id, f); len(sid) > 0 {
						f.SetUrl(profile.UrlToPattern(u), nil, &profile.UrlProxyAction{profile.UrlActRestore, sid})
						fmt.Fprintf(w, "<html><head><title>缓存历史 %s</title></head><body>历史 <a href=\"/profile/%s/stores/%s\">%s</a> 已缓存至 URL %s。<br/>返回 <a href=\"/profile/%s\">管理页面</a></body></html>", id, profileIP, sid, id, u, profileIP)
					}
					return
				}
			}
		}

		fmt.Fprintf(w, "<html><body>无效请求。返回 <a href=\"/profile/%s\">管理页面</a></body></html>", profileIP)
		return
	} else if op == "stores" {
		if len(pages) >= 4 {
			k := pages[3]
			switch k {
			case "edit":
				id := ""
				if len(pages) >= 5 {
					id = pages[4]
				}
				r.ParseForm()
				if v, ok := r.Form["id"]; ok && len(v) > 0 {
					id = strings.TrimSpace(v[0])
					if space := strings.Index(id, " "); space >= 0 {
						id = id[:space]
					}
				}

				if v, ok := r.Form["content"]; ok && len(v) > 0 {
					if c, err := url.QueryUnescape(v[0]); err == nil {
						f.Store(id, []byte(c))
					}
				}

				p.writeEditStore(w, profileIP, f, id)
			default:
				c := f.Restore(k)
				if len(c) > 0 {
					w.Write(c)
				} else {
					w.WriteHeader(404)
				}
			}
			return
		} else {
			p.writeStores(w, profileIP, f)
		}
		return
	} else if op != "" {
		fmt.Fprintf(w, "<html><body>无效请求 %s。<br/>返回 <a href=\"/profile/%s\">管理页面</a></body></html>", op, profileIP)
		return
	}

	r.ParseForm()
	if v, ok := r.Form["cmd"]; ok && len(v) > 0 {
		if f.Owner == ownerIP || profileIP == ownerIP {
			for _, cmd := range v {
				p.Command(cmd, f, p.lives.Open(profileIP))
			}
		}
	}

	savedIDs := f.ListStoreIDs()
	f.WriteHtml(w, savedIDs)
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

func (p *Proxy) storeHistory(profileIP, id string, prof *profile.Profile) (string, string) {
	f := p.lives.OpenExists(profileIP)
	if f == nil {
		return "", ""
	}

	hID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return "", ""
	}

	h := f.LookHistoryByID(uint32(hID))
	if h.Error == nil && len(h.Bytes) > 0 {
		saveID := prof.StoreID(h.Bytes)
		return h.Url, saveID
	}

	return "", ""
}

func (p *Proxy) dns(page string, w http.ResponseWriter, r *http.Request) {
	f := p.profileOp.Open("localhost")
	if f == nil {
		fmt.Fprintln(w, "无效")
		return
	}

	r.ParseForm()
	if v, ok := r.Form["cmd"]; ok && len(v) > 0 {
		for _, cmd := range v {
			p.Command(cmd, f, nil)
		}
	}

	if len(page) == 0 {
		f.WriteDNS(w)
	} else if page == "/export" {
		export := "# 此为 DNS 独立服务的配置导出，可复制所有内容至“命令”输入窗口重新加载此配置 #\n\n"
		export += "# Name: DNS 独立服务\n"
		export += f.ExportDNSCommand()
		export += "\n# end #\n"
		fmt.Fprintln(w, export)
	}
}

func (p *Proxy) isLoopback(addr string) bool {
	ip := gonet.ParseIP(addr)
	if ip != nil && ip.IsLoopback() {
		return true
	}

	return strings.EqualFold(addr, "localhost")
}

func (p *Proxy) isSelfAddr(addr string) bool {
	if p.isLoopback(addr) {
		return true
	} else if addr == p.serveIP {
		return true
	}

	return false
}
