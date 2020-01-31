package proxy

import (
	"github.com/benbearchen/asuran/net"
	"github.com/benbearchen/asuran/net/httpd"
	"github.com/benbearchen/asuran/net/websocket"
	"github.com/benbearchen/asuran/policy"
	"github.com/benbearchen/asuran/profile"
	"github.com/benbearchen/asuran/web/proxy/cache"
	"github.com/benbearchen/asuran/web/proxy/life"
	"github.com/benbearchen/asuran/web/proxy/pack"
	_ "github.com/benbearchen/asuran/web/proxy/plugin"
	"github.com/benbearchen/asuran/web/proxy/plugin/api"
	_ "github.com/benbearchen/asuran/web/proxy/tunnel"
	tunnel "github.com/benbearchen/asuran/web/proxy/tunnel/api"

	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	gonet "net"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ASURAN_POLICY_HEADER = "ASURAN_POLICY"
	ASURAN_PACK_HEADER   = "ASURAN_PACK"
)

type Proxy struct {
	ver        string
	webServers map[int]*httpd.Http
	lives      *life.IPLives
	urlOp      profile.UrlOperator
	profileOp  profile.ProfileOperator
	domainOp   profile.DomainOperator
	serveIP    string
	mainHost   string
	domain     string
	proxyAddr  string
	disableDNS bool
	packs      *pack.Dir

	lock sync.RWMutex
	r    *rand.Rand
}

func NewProxy(ver string, dataDir string) *Proxy {
	p := new(Proxy)
	p.ver = ver
	p.webServers = make(map[int]*httpd.Http)
	p.lives = life.NewIPLives()
	p.r = rand.New(rand.NewSource(time.Now().UnixNano()))
	p.packs = pack.New(filepath.Join(dataDir, "packs"))
	p.domain = "asu.run"

	p.Bind(80)

	ips := net.LocalIPs()
	if ips == nil {
		fmt.Println("大王不好了，找不到本地 IP 啊！！")
	} else {
		for _, ip := range ips {
			p.serveIP = ip
			fmt.Println("HTTP 代理、DNS 服务 监听 IP: " + ip)
			fmt.Println()
			for port, _ := range p.webServers {
				p.mainHost = ip
				if port != 80 {
					p.mainHost += ":" + strconv.Itoa(port)
				}

				p.proxyAddr = ip + ":" + strconv.Itoa(port)

				fmt.Println("标准 HTTP 代理地址:  " + p.proxyAddr)
				fmt.Println("asuran 管理界面:     http://" + p.mainHost + "/    ∈←← ←  ←    ←")
			}

			fmt.Println()
		}

		fmt.Println("本机还可访问 http://localhost/ 进入管理界面")
	}

	return p
}

func (p *Proxy) SetVisitDomain(domain string) {
	p.domain = domain
}

func (p *Proxy) Bind(port int) bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, exists := p.webServers[port]; exists {
		return false
	}

	h := &httpd.Http{}
	p.webServers[port] = h
	serverAddress := fmt.Sprintf(":%d", port)
	h.Init(serverAddress)
	h.RegisterHandler(p)
	go h.Run(func(err error) { p.overHttpd(port, err) })
	return true
}

func (p *Proxy) overHttpd(port int, err error) {
	if err == nil {
		fmt.Println("bind on port", port, "quit")
	} else {
		fmt.Println("bind on port", port, "failed with:", err)
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.webServers, port)
}

func (p *Proxy) TryBind(port int) {
	p.Bind(port)
}

func (p *Proxy) BindUrlOperator(op profile.UrlOperator) {
	p.urlOp = op
}

func (p *Proxy) BindProfileOperator(op profile.ProfileOperator) {
	p.profileOp = op
	p.profileOp.Open("localhost").Name = "DNS 服务"
	p.lives.Open("localhost")
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

func (p *Proxy) DisableDNS() {
	p.disableDNS = true
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

func (p *Proxy) OnRequest(w http.ResponseWriter, r *http.Request) {
	targetHost := httpd.RemoteHost(r.Host)
	remoteIP := httpd.RemoteHost(r.RemoteAddr)
	urlPath := r.URL.Path
	if u, ok := r.Header["Upgrade"]; ok {
		if len(u) > 0 && strings.ToLower(u[0]) == "websocket" {
			p.proxyWebsocket(remoteIP, w, r)
			return
		}
	}

	//fmt.Printf("host: %s/%s, remote: %s/%s, url: %s\n", targetHost, r.Host, remoteIP, r.RemoteAddr, urlPath)
	if r.Method != "GET" && r.Method != "POST" {
		w.WriteHeader(502)
		fmt.Fprintln(w, "unknown method", r.Method, "to", r.Host)
	} else if strings.HasPrefix(urlPath, "http://") {
		p.proxyRequest(w, r)
	} else if targetHost == p.domain {
		p.initDevice(w, remoteIP)
	} else if !p.isSelfAddr(targetHost) && !p.isSelfAddr(remoteIP) {
		p.proxyRequest(w, r)
	} else if _, m := httpd.MatchPath(urlPath, "/post"); m {
		p.postTest(w, r)
	} else if page, m := httpd.MatchPath(urlPath, "/test/"); m {
		if page == "" {
			page = "localhost"
		}

		target := "http://" + page[1:]
		p.testUrl(target, w, r)
	} else if page, m := httpd.MatchPath(urlPath, "/to/"); m {
		target := "http://" + page[1:]
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}

		p.proxyUrl(target, w, r)
	} else if _, m := httpd.MatchPath(urlPath, "/usage"); m {
		p.WriteUsage(w)
	} else if page, m := httpd.MatchPath(urlPath, "/profile"); m {
		p.ownProfile(remoteIP, page, w, r)
	} else if page, m := httpd.MatchPath(urlPath, "/dns"); m {
		if !p.disableDNS {
			p.dns(page, w, r)
		} else {
			fmt.Fprintln(w, "DNS is disabled")
		}
	} else if _, m := httpd.MatchPath(urlPath, "/res"); m {
		p.res(w, r, urlPath)
	} else if page, m := httpd.MatchPath(urlPath, "/packs"); m {
		p.dealPacks(w, r, page)
	} else if page, m := httpd.MatchPath(urlPath, "/plugins"); m {
		p.dealPlugins(w, r, page)
	} else if page, m := httpd.MatchPath(urlPath, "/tunnel"); m {
		p.tunnel(w, r, page)
	} else if urlPath == "/" {
		ip := func() string {
			if p.isSelfAddr(remoteIP) {
				return ""
			} else {
				return remoteIP
			}
		}()

		accessCode := ""
		if len(ip) > 0 {
			prof := p.profileOp.FindByIp(ip)
			if prof != nil {
				accessCode = prof.AccessCode()
			}
		}

		p.index(w, p.ver, ip, accessCode)
	} else if urlPath == "/features" {
		p.features(w, p.ver)
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
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "visit http://"+p.mainHost+"/profile/255.255.255.255/policy/speed%2015/"+p.mainHost+"/about to test the command policy")
	} else if urlPath == "/urlencoded" {
		p.urlEncoded(w)
	} else {
		fmt.Fprintln(w, "visit http://"+r.Host+"/about to get info")
	}
}

func (p *Proxy) proxyRequest(w http.ResponseWriter, r *http.Request) {
	scheme := r.URL.Scheme
	if len(scheme) == 0 {
		scheme = "http"
	}

	url := scheme + "://" + r.Host + r.URL.RequestURI()
	p.proxyUrl(url, w, r)
}

func (p *Proxy) proxyUrl(target string, w http.ResponseWriter, r *http.Request) {
	remoteIP := httpd.RemoteHost(r.RemoteAddr)
	p.remoteProxyUrl(remoteIP, target, w, r, nil)
}

func (p *Proxy) remoteProxyUrl(remoteIP, target string, w http.ResponseWriter, r *http.Request, up *policy.UrlPolicy) {
	needCache := false

	fullUrl := target
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
		in := f.Incoming(fullUrl, "")
		defer in.Done()
	}

	var writeWrap io.Writer = nil
	forceChunked := false
	forceRecvFirst := false

	if up == nil {
		if cmd := r.Header.Get(ASURAN_POLICY_HEADER); len(cmd) > 0 {
			p, err := policy.Factory("url " + cmd)
			if err != nil {
				w.WriteHeader(400)
				fmt.Fprintf(w, `policy "url %s" err: %v`, cmd, err)
				return
			}

			r.Header.Del(ASURAN_POLICY_HEADER)
			up = p.(*policy.UrlPolicy)
		} else if pkg := r.Header.Get(ASURAN_PACK_HEADER); len(pkg) > 0 {
			p, err := p.matchPack(fullUrl, pkg)
			if err != nil {
				w.WriteHeader(400)
				fmt.Fprintf(w, `policy pack "%s" err: %v`, pkg, err)
				return
			}

			r.Header.Del(ASURAN_PACK_HEADER)
			up = p
		} else if p.urlOp != nil {
			up = p.urlOp.Action(remoteIP, fullUrl)
		}
	}

	if up != nil && up.Plugin() != nil {
		p.plugin(remoteIP, up, target, w, r, f)
		return
	}

	if up != nil {
		delay := up.DelayPolicy()
		if delay != nil {
			//fmt.Println("url delay: " + delay.String())
			switch delay := delay.(type) {
			case *policy.DelayPolicy:
				if delay.Duration() > 0 {
					// TODO: create request before sleep, more effective
					d := delay.RandDuration(p.r)
					time.Sleep(d)
					f.Log("proxy " + fullUrl + " delay " + d.String())
				}
				break
			case *policy.DropPolicy:
				d := delay.RandDuration(p.r)
				if u != nil && u.DropUntil(d) {
					f.Log("proxy " + fullUrl + " drop " + d.String())
					net.ResetResponse(w)
					return
				}
				break
			case *policy.TimeoutPolicy:
				if delay.Duration() > 0 {
					d := delay.RandDuration(p.r)
					time.Sleep(d)
					f.Log("proxy " + fullUrl + " timeout " + d.String())
				} else {
					f.Log("proxy " + fullUrl + " timeout")
				}
				net.ResetResponse(w)
				return
				break
			}
		}

		if s := up.Status(); s != nil {
			status := s.StatusCode()
			if status == 0 {
				status = 502
			}

			w.WriteHeader(status)
			f.Log("proxy " + fullUrl + " status " + strconv.Itoa(status))
			return
		}

		act := up.ContentPolicy()
		//fmt.Println("url act: " + act.String())
		speed := up.Speed()
		bodyDelay := up.BodyPolicy()
		chunked := up.Chunked()

		if act != nil {
			switch act := act.(type) {
			case *policy.CachePolicy:
				needCache = true
			case *policy.MapPolicy:
				requestUrl = act.URL(requestUrl)
				requestR = nil
				contentSource = "map " + requestUrl
			case *policy.RedirectPolicy:
				requestUrl = act.URL(requestUrl)
				http.Redirect(w, r, requestUrl, 302)
				f.Log("proxy " + fullUrl + " redirect " + requestUrl)
				return
			case *policy.RewritePolicy, *policy.RestorePolicy, *policy.TcpwritePolicy:
				if p.rewriteUrl(fullUrl, w, r, rangeInfo, prof, f, act, speed, chunked, bodyDelay, up.ContentType(), up.ResponseHeaders()) {
					return
				}
			}
		}

		if chunked != nil {
			chunkedOp := chunked.Option()
			if chunkedOp != policy.ChunkedDefault && chunkedOp != policy.ChunkedOff {
				forceChunked = true
			}

			if chunkedOp == policy.ChunkedOff || chunkedOp == policy.ChunkedBlock || chunkedOp == policy.ChunkedSize {
				forceRecvFirst = true
			}
		}

		if bodyDelay != nil {
			switch bodyDelay.(type) {
			case *policy.DelayPolicy, *policy.TimeoutPolicy:
				if writeWrap == nil {
					writeWrap = w
				}

				canSubPackage := !forceChunked
				writeWrap = newDelayWriter(bodyDelay, writeWrap, p.r, canSubPackage)
			}
		}

		if speed != nil {
			if writeWrap == nil {
				writeWrap = w
			}

			canSubPackage := !forceChunked
			writeWrap = newSpeedWriter(speed, writeWrap, canSubPackage)
		}

		if forceChunked {
			if writeWrap == nil {
				writeWrap = w
			}

			writeWrap = newChunkedWriter(chunked, writeWrap)
		}
	}

	if needCache && r.Method == "GET" && f != nil {
		c := f.CheckCache(fullUrl, rangeInfo)
		if c != nil && c.Error == nil {
			c.Response(w, writeWrap)
			extraInfo := ""
			if len(rangeInfo) > 0 {
				extraInfo = " " + rangeInfo
			}

			f.Log("cache " + fullUrl + extraInfo)
			return
		}
	}

	dont302 := true
	settingContentType := "default"
	var hostPolicy *policy.HostPolicy
	var headersPolicy *policy.HeadersPolicy
	if up != nil {
		dont302 = up.Dont302()
		settingContentType = up.ContentType()
		if requestR != nil {
			if up.Disable304() {
				p.disable304FromHeader(requestR.Header)
			}

			hp := up.RequestHeaders()
			if hp != nil {
				hp.Apply(requestR.Header)
			}
		}

		hostPolicy = up.Host()
		headersPolicy = up.ResponseHeaders()
	}

	httpStart := time.Now()
	resp, postBody, redirection, err := net.NewHttp(requestUrl, requestR, p.parseDomainAsDial(requestUrl, remoteIP, hostPolicy), dont302)
	if err != nil {
		c := cache.NewUrlCache(fullUrl, r, nil, nil, contentSource, nil, rangeInfo, httpStart, time.Now(), err)
		if f != nil {
			go p.saveContentToCache(fullUrl, f, c, false)
		}

		http.Error(w, "Bad Gateway", 502)
	} else if len(redirection) > 0 {
		http.Redirect(w, r, redirection, 302)
		if f != nil {
			f.Log("proxy " + fullUrl + " redirect " + redirection)
		}
	} else {
		defer resp.Close()
		p.procHeader(resp.Header(), settingContentType, headersPolicy)
		content, err := resp.ProxyReturn(w, writeWrap, forceRecvFirst, forceChunked)
		httpEnd := time.Now()
		c := cache.NewUrlCache(fullUrl, r, postBody, resp, contentSource, content, rangeInfo, httpStart, httpEnd, err)
		if f != nil {
			go p.saveContentToCache(fullUrl, f, c, needCache)
		}
	}
}

func (p *Proxy) rewriteUrl(target string, w http.ResponseWriter, r *http.Request, rangeInfo string, prof *profile.Profile, f *life.Life, act policy.Policy, speed *policy.SpeedPolicy, chunked *policy.ChunkedPolicy, bodyDelay policy.Policy, contentType string, hp *policy.HeadersPolicy) bool {
	var content []byte = nil
	contentSource := ""
	istcp := false
	switch act := act.(type) {
	case *policy.RewritePolicy:
		u, err := url.QueryUnescape(act.Value())
		if err != nil {
			return false
		}

		content = []byte(u)
		contentSource = "rewrite"
	case *policy.TcpwritePolicy:
		istcp = true
		u, err := url.QueryUnescape(act.Value())
		if err != nil {
			return false
		}

		content = []byte(u)
		contentSource = "tcpwrite"
	case *policy.RestorePolicy:
		content = prof.Restore(act.Value())
		if content == nil {
			return false
		}
	default:
		return false
	}

	if len(rangeInfo) > 0 {
		c, cr, err := cache.MakeRange(rangeInfo, content)
		if err != nil {
			w.WriteHeader(416)
			fmt.Fprintln(w, "error:", err)
			return true
		} else {
			w.Header()["Content-Range"] = []string{"bytes " + cr}
			w.Header()["Content-Length"] = []string{strconv.Itoa(len(c))}
			w.WriteHeader(206)
			content = c
		}
	}

	if _, ok := act.(*policy.TcpwritePolicy); !ok {
		p.procHeader(w.Header(), contentType, hp)
	}

	forceChunked := false
	if chunked != nil {
		chunkedOp := chunked.Option()
		if chunkedOp != policy.ChunkedDefault && chunkedOp != policy.ChunkedOff {
			forceChunked = true
		}
	}

	var writeWrapper func(w io.Writer) io.Writer = nil
	if bodyDelay != nil {
		switch bodyDelay.(type) {
		case *policy.DelayPolicy, *policy.TimeoutPolicy:
			writeWrapper = func(w io.Writer) io.Writer {
				return newDelayWriter(bodyDelay, w, p.r, true)
			}
		}
	}

	if speed != nil {
		canSubPackage := !forceChunked
		if writeWrapper != nil {
			wrap := writeWrapper
			writeWrapper = func(w io.Writer) io.Writer {
				return newSpeedWriter(speed, wrap(w), canSubPackage)
			}
		} else {
			writeWrapper = func(w io.Writer) io.Writer {
				return newSpeedWriter(speed, w, canSubPackage)
			}
		}
	}

	if forceChunked {
		if writeWrapper != nil {
			wrap := writeWrapper
			writeWrapper = func(w io.Writer) io.Writer {
				return newChunkedWriter(chunked, wrap(w))
			}
		} else {
			writeWrapper = func(w io.Writer) io.Writer {
				return newChunkedWriter(chunked, w)
			}
		}
	}

	start := time.Now()
	if istcp {
		net.TcpWriteHttp(w, writeWrapper, content)
	} else {
		if !forceChunked {
			// set the Content-Length, then chunked would be disabled
			w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		}

		w.WriteHeader(200)
		if writeWrapper != nil {
			writeWrapper(w).Write(content)
		} else {
			w.Write(content)
		}
	}

	c := cache.NewUrlCache(target, r, nil, nil, contentSource, content, rangeInfo, start, time.Now(), nil)
	if istcp {
		c.ResponseCode = 599
	} else {
		c.ResponseCode = 200
	}
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

func (p *proxyDomainOperator) Action(ip, domain string) *policy.DomainPolicy {
	if domain == p.p.domain {
		p.p.LogDomain(ip, ip, "init", domain, p.p.serveIP)
		return policy.NewStaticDomainPolicy(domain, p.p.serveIP)
	}

	profIP := ip
	if p.p.profileOp != nil {
		if p.p.profileOp.FindByIp(ip) == nil {
			profIP = "localhost"
		}
	}

	if p.p.domainOp != nil {
		act := "query"
		a := p.p.domainOp.Action(profIP, domain)
		if a != nil && a.Action() != nil {
			switch a.Action().(type) {
			case *policy.ProxyPolicy:
				a = policy.NewStaticDomainPolicy(domain, p.p.serveIP)
				act = "proxy"
			case *policy.BlockPolicy:
				act = "block"
			case *policy.NullPolicy:
				act = "null"
			}
		}

		resultIP := ""
		if a != nil && len(a.IP()) > 0 {
			resultIP = a.IP()
		}

		p.p.LogDomain(profIP, ip, act, domain, resultIP)
		return a
	} else {
		p.p.LogDomain(profIP, ip, "undef", domain, "")
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

	canOperate := f.CanOperate(ownerIP)

	if op == "export" {
		fmt.Fprintln(w, f.ExportCommand())
		return
	} else if op == "export1" || op == "export2" || op == "export3" {
		i, err := strconv.Atoi(op[6:])
		if err != nil {
			w.WriteHeader(403)
			fmt.Fprintln(w, "无效参数：", op)
		} else {
			cmd, err := f.ExportHistoryCommand(i - 1)
			if err != nil {
				w.WriteHeader(403)
				fmt.Fprintln(w, err)
			} else {
				fmt.Fprintln(w, cmd)
			}
		}

		return
	} else if op == "restart" {
		if !canOperate {
			w.WriteHeader(403)
			fmt.Fprintln(w, "没有操作权限")
		} else if f := p.lives.Visit(profileIP); f != nil {
			f.Restart()
			p.resetPlugin(profileIP)
			fmt.Fprintf(w, "restarted")
		} else {
			w.WriteHeader(404)
			fmt.Fprintln(w, profileIP+" 不存在")
		}

		return
	} else if op == "history" {
		if f := p.lives.Visit(profileIP); f != nil {
			if len(pages) >= 4 && pages[3] == "watch.json" {
				p.watchHistory(w, r, profileIP, f)
			} else if len(pages) >= 4 && pages[3] == "clear" {
				f.ClearHistory()
				fmt.Fprintf(w, "cleared")
			} else {
				p.writeHistory(w, profileIP, f)
			}
		} else {
			fmt.Fprintln(w, profileIP+" 不存在")
		}
		return
	} else if op == "in.json" {
		if f := p.lives.OpenExists(profileIP); f != nil {
			p.watchIncoming(w, r, profileIP, f)
		} else {
			w.WriteHeader(404)
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
					dp, _ := policy.Factory("domain proxy " + domain)
					f.SetDomainPolicy(dp.(*policy.DomainPolicy))
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
						up, _ := policy.Factory("url restore " + sid + " " + profile.UrlToPattern(u))
						f.SetUrlPolicy(up.(*policy.UrlPolicy), nil, nil)
						p.writeStoreResult(w, profileIP, u, id, sid)
					}
					return
				}
			}
		}

		fmt.Fprintf(w, "<html><body>无效请求。返回 <a href=\"/profile/%s\">管理页面</a></body></html>", profileIP)
		return
	} else if op == "stores" {
		if len(pages) >= 4 && pages[3] != "" {
			k := pages[3]
			switch k {
			case "edit", "view":
				id := ""
				if len(pages) >= 5 {
					id = pages[4]
				}

				p.writeEditStore(w, profileIP, f, id, k == "view")
			case "commit":
				if !canOperate {
					w.WriteHeader(403)
					fmt.Fprintln(w, "无权操作")
					return
				}

				id := ""

				r.ParseForm()
				if v, ok := r.Form["id"]; ok && len(v) > 0 {
					id = strings.TrimSpace(v[0])
					if space := strings.Index(id, " "); space >= 0 {
						id = id[:space]
					}
				}

				if id == "" {
					w.WriteHeader(404)
					fmt.Fprintln(w, "empty id")
					return
				}

				if v, ok := r.Form["content"]; ok && len(v) > 0 {
					c, err := url.QueryUnescape(v[0])
					if err == nil {
						f.Store(id, []byte(c))
						fmt.Fprintf(w, "保存成功")
					} else {
						w.WriteHeader(404)
						fmt.Fprintf(w, "错误:%v", err)
					}

					return
				}

				w.WriteHeader(404)
				fmt.Fprintln(w, "empty content")
			case "delete":
				if len(pages) >= 5 {
					id := pages[4]
					f.DeleteStore(id)
					fmt.Fprintln(w, "已删除 "+id)
				} else {
					fmt.Fprintln(w, "请指定 Store ID")
				}
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
	} else if op == "operator" {
		if !canOperate {
			if len(pages) >= 5 && pages[3] == "apply" {
				if f.CheckAccessCode(pages[4]) {
					f.AddOperator(ownerIP)
					fmt.Fprintf(w, pages[4])
				} else {
					w.WriteHeader(403)
					fmt.Fprintf(w, "访问码错误")
				}
			} else {
				w.WriteHeader(403)
				fmt.Fprintf(w, "<html><body>无权操作 %s。<br/>返回 <a href=\"/profile/%s\">管理页面</a></body></html>", profileIP, profileIP)
			}

			return
		} else {
			if len(pages) >= 5 {
				operator := pages[4]
				switch pages[3] {
				case "add":
					f.AddOperator(operator)
				case "remove":
					f.RemoveOperator(operator)
				default:
					w.WriteHeader(404)
					fmt.Fprintf(w, "未知操作 %s", pages[3])
					return
				}
			}

			operators := ""
			for op, _ := range f.Operators {
				if len(operators) == 0 {
					operators = op
				} else {
					operators += ", " + op
				}
			}

			fmt.Fprintf(w, "%s", operators)
		}
		return
	} else if op == "to" {
		if len(pages) <= 3 || pages[3] == "" {
			w.WriteHeader(400)
			fmt.Fprintln(w, "profile/.../to/ need a target, like profile/.../to/g.cn")
			return
		}

		url := "http://" + strings.Join(pages[3:], "/")
		if r.URL.RawQuery != "" {
			url = url + "?" + r.URL.RawQuery
		}

		p.remoteProxyUrl(profileIP, url, w, r, nil)
		return
	} else if op == "policy" {
		if len(pages) <= 4 || pages[4] == "" {
			w.WriteHeader(400)
			fmt.Fprintln(w, "profile/.../policy/ need policy command and target, like profile/.../to/rewrite xyz/g.cn")
			return
		}

		cmd := pages[3]
		up, err := policy.Factory("url " + cmd)
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, `policy "url %s" err: %v`, cmd, err)
			return
		}

		url := "http://" + strings.Join(pages[4:], "/")
		if r.URL.RawQuery != "" {
			url = url + "?" + r.URL.RawQuery
		}

		p.remoteProxyUrl(profileIP, url, w, r, up.(*policy.UrlPolicy))
		return
	} else if op == "match" {
		r.ParseForm()
		if v, ok := r.Form["url"]; ok && len(v) > 0 {
			c := f.UrlAction(v[0])
			if c != nil {
				fmt.Fprintf(w, "%s", c.Policy())
			}
		} else {
			w.WriteHeader(404)
			fmt.Fprintln(w, "miss url")
		}

		return
	} else if op == "pattern" {
		p.patternProc(w, r)
		return
	} else if op == "pack" {
		p.packCommand(w, r)
		return
	} else if op == "check" {
		r.ParseForm()
		errors := []string{}
		if v, ok := r.Form["cmd"]; ok && len(v) > 0 {
			errors = p.CheckCommand(v[0])
			for _, e := range errors {
				fmt.Fprintln(w, e)
			}
		} else {
			w.WriteHeader(403)
		}

		return
	} else if op != "" {
		w.WriteHeader(404)
		fmt.Fprintf(w, "<html><body>无效请求 %s。<br/>返回 <a href=\"/profile/%s\">管理页面</a></body></html>", op, profileIP)
		return
	}

	p.lives.Visit(profileIP)

	r.ParseForm()
	errors := []string{}
	if v, ok := r.Form["cmd"]; ok && len(v) > 0 {
		if canOperate {
			for _, cmd := range v {
				errors = p.Command(cmd, f, p.lives.Open(profileIP))
			}
		}
	}

	savedIDs := f.ListStoreIDs()
	f.WriteHtml(w, savedIDs, canOperate, errors)
}

func (p *Proxy) lookHistoryByID(w http.ResponseWriter, profileIP string, id uint32, op string) {
	f := p.lives.Visit(profileIP)
	if f == nil {
		fmt.Fprintln(w, profileIP+" 不存在")
		return
	}

	h := f.LookHistoryByID(id)
	if h != nil {
		switch op {
		case "look":
			h.Response(w, nil)
		case "detail":
			h.Detail(w)
		}
	} else {
		fmt.Fprintf(w, "history %d not exist", id)
	}
}

func (p *Proxy) lookHistory(w http.ResponseWriter, profileIP, lookUrl, op string) {
	f := p.lives.Visit(profileIP)
	if f == nil {
		fmt.Fprintln(w, profileIP+" 不存在")
		return
	}

	switch op {
	case "look":
		if c := f.LookCache(lookUrl); c != nil {
			c.Response(w, nil)
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

func (p *Proxy) LogDomain(profIP, ip, action, domain, resultIP string) {
	if f := p.lives.Open(profIP); f != nil {
		s := ""
		if len(resultIP) > 0 {
			s = " "
		}

		d := "domain"
		if profIP == "localhost" {
			d = ip
		}

		f.Log(d + " " + action + " " + domain + s + resultIP)
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

func (p *Proxy) parseDomainAsDial(target, client string, hostPolicy *policy.HostPolicy) func(network, addr string) (gonet.Conn, error) {
	address := ""
	if hostPolicy == nil {
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
		if a == nil || len(a.IP()) == 0 {
			return nil
		}

		address = gonet.JoinHostPort(a.IP(), port)
	} else {
		address = hostPolicy.HTTP()
	}

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
	if h == nil {
		return "", ""
	}

	content, err := h.Content()
	if err == nil {
		saveID := prof.StoreID(content)
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

	if len(page) == 0 || page == "/" {
		f.WriteDNS(w, p.serveIP)
	} else if _, m := httpd.MatchPath(page, "/export"); m {
		export := "# 此为 DNS 独立服务的配置导出，可复制所有内容至“命令”输入窗口重新加载此配置 #\n\n"
		export += "# Name: DNS 独立服务\n"
		export += f.ExportDNSCommand()
		export += "\n# end #\n"
		fmt.Fprintln(w, export)
	} else if target, m := httpd.MatchPath(page, "/history"); m {
		if len(target) > 0 {
			target = target[1:]
		}

		p.writeDNSHistory(w, p.lives.Open("localhost"), target)
	} else {
		http.Redirect(w, r, "..", 302)
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

func (p *Proxy) procHeader(header http.Header, settingContentType string, hp *policy.HeadersPolicy) {
	switch settingContentType {
	case "default":
	case "remove":
		header["Content-Type"] = nil
	case "empty":
		settingContentType = ""
		fallthrough
	default:
		header["Content-Type"] = []string{settingContentType}
	}

	if hp != nil {
		hp.Apply(header)
	}
}

func (p *Proxy) disable304FromHeader(header http.Header) {
	delete(header, "If-None-Match")
	delete(header, "If-Modified-Since")
}

func (p *Proxy) proxyWebsocket(client string, w http.ResponseWriter, r *http.Request) {
	//fmt.Printf("recv websocket: to %s %v\n", r.Host, r.URL)
	domain, port, err := gonet.SplitHostPort(r.Host)
	if err != nil {
		domain = r.Host
	}

	if len(port) == 0 {
		port = "80"
	}

	host := domain
	if p.domainOp != nil {
		a := p.domainOp.Action(client, domain)
		if a != nil && len(a.IP()) > 0 {
			host = a.IP()
		}
	}

	address := gonet.JoinHostPort(host, port)

	headers := make(map[string][]string)
	for h, v := range r.Header {
		headers[h] = v
	}

	if _, ok := headers["Host"]; !ok {
		headers["Host"] = []string{r.Host}
	}

	path := r.URL.RequestURI()

	upConn, err := websocket.Conn(address, path, headers)
	if err != nil {
		w.WriteHeader(502)
		fmt.Fprintln(w, err)
		return
	}

	downConn, _, err := net.TryHijack(w)
	if err != nil {
		w.WriteHeader(502)
		fmt.Fprintln(w, err)
		return
	}

	net.PipeConn(upConn, downConn)
}

func (p *Proxy) packCommand(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.Form.Get("name")
	author := r.Form.Get("author")
	comment := r.Form.Get("comment")
	cmd := r.Form.Get("cmd")

	if len(name) == 0 || len(author) == 0 || len(cmd) == 0 {
		fmt.Fprintf(w, "error: some fields are empty")
		return
	}

	err := p.packs.Save(name, author, comment, cmd)
	if err != nil {
		fmt.Fprintf(w, "save failed, err: %v", err)
	} else {
		fmt.Fprintf(w, "saved")
	}
}

func (p *Proxy) dealPacks(w http.ResponseWriter, r *http.Request, page string) {
	if len(page) <= 1 { // "" or "/"
		p.writePacks(w)
		return
	}

	switch page[1:] {
	case "names.json":
		bytes, err := json.Marshal(p.packs.ListNames())
		if err == nil {
			w.Write(bytes)
		} else {
			w.WriteHeader(502)
			fmt.Fprintf(w, "error: %v", err)
		}
		return
	case "get":
		r.ParseForm()
		cmd := p.packs.Get(r.Form.Get("name"))
		if len(cmd) > 0 {
			fmt.Fprintf(w, "%s", cmd)
		} else {
			w.WriteHeader(404)
			fmt.Fprintln(w, "invalid name or pack is empty")
		}
		return
	}

	w.WriteHeader(404)
}

func (p *Proxy) dealPlugins(w http.ResponseWriter, r *http.Request, page string) {
	if len(page) <= 1 {
		p.writePlugins(w)
		return
	}

	switch page[1:] {
	case "names.json":
		bytes, err := json.Marshal(api.All())
		if err == nil {
			w.Write(bytes)
		} else {
			w.WriteHeader(502)
			fmt.Fprintf(w, "error: %v", err)
		}
		return
	case "intro":
		r.ParseForm()
		intro := api.Intro(r.Form.Get("name"))
		fmt.Fprintf(w, "%s", intro)
		return
	}

	w.WriteHeader(404)
}

func (p *Proxy) tunnel(w http.ResponseWriter, r *http.Request, page string) {
	if len(page) <= 1 {
		p.writeTunnels(w)
		return
	}

	name, path := httpd.PopPath(page)
	tun := tunnel.Get(name)
	if tun != nil {
		url := tun.Link() + path
		if r.URL.RawQuery != "" {
			url = url + "?" + r.URL.RawQuery
		}

		p.proxyUrl(url, w, r)
	} else {
		w.WriteHeader(404)
		fmt.Fprintln(w, "invalid tunnel:", name)
	}
}

func (p *Proxy) tunnelPrefix() string {
	return "http://" + p.mainHost + "/tunnel"
}

func (*Proxy) patternProc(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	op := r.Form.Get("op")
	if op != "test" {
		w.WriteHeader(404)
		fmt.Fprintln(w, "unknown operation: "+op)
		return
	}

	t := r.Form.Get("t")
	p := r.Form.Get("pattern")
	v := r.Form.Get("v")

	var err error
	match := false
	switch t {
	case "domain":
		match = profile.NewDomainPattern(p).Match(v)
	case "path":
		match = profile.NewPathPattern(p).Match(v)
	case "args":
		match = profile.NewArgsPattern(p).MatchArgs(v)
	case "url":
		match = profile.NewUrlPattern(p).MatchUrl(v)
	default:
		w.WriteHeader(404)
		fmt.Fprintln(w, "unknown pattern type: "+r.Form.Get("t"))
		return
	}

	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, "err: %v", err)
		return
	}

	result := "成功"
	if !match {
		result = "失败"
	}

	fmt.Fprintf(w, "%s", result)
}

func (p *Proxy) plugin(profileIP string, up *policy.UrlPolicy, target string, w http.ResponseWriter, r *http.Request, f *life.Life) {
	start := time.Now()
	pluginPolicy := up.Plugin()
	log := func(statusCode int, postBody, content []byte, err error) {
		go func() {
			c := cache.NewUrlCache(target, r, postBody, nil, "plugin "+pluginPolicy.Name(), content, "", start, time.Now(), err)
			c.ResponseCode = statusCode

			id := f.SaveContentToCache(c, false)

			info := "plugin " + target + " " + strconv.FormatUint(uint64(id), 10) + " " + pluginPolicy.Name()

			f.Log(info)
		}()
	}

	context := &policy.PluginContext{profileIP, up.Target(), log}
	api.Call(context, pluginPolicy.Name(), w, r)
}

func (p *Proxy) resetPlugin(profileIP string) {
	context := &policy.PluginContext{ProfileIP: profileIP}
	api.Reset(context, "")
}

func (p *Proxy) matchPack(fullUrl, packName string) (*policy.UrlPolicy, error) {
	cmd := p.packs.Get(packName)
	if len(cmd) == 0 {
		return nil, fmt.Errorf("has no pack %s", packName)
	}

	ps, _ := p.ParseCommand(cmd)
	var score uint32 = 0
	up := []*policy.UrlPolicy{}
	for _, p := range ps {
		switch p := p.(type) {
		case *policy.UrlPolicy:
			var s uint32 = 0
			if p.Target() != "" {
				s = profile.NewUrlPattern(p.Target()).MatchUrlScore(fullUrl)
			}

			if s > score {
				score = s
				up = []*policy.UrlPolicy{p}
			} else if s == score {
				up = append(up, p)
			}
		default:
			continue
		}
	}

	if len(up) > 0 {
		return up[rand.Int()%len(up)], nil
	} else {
		return nil, nil
	}
}

type jsonWatchHistory struct {
	Info    string             `json:"info"`
	History []historyEventData `json:"history"`
}

func (p *Proxy) watchHistory(w http.ResponseWriter, r *http.Request, profileIP string, f *life.Life) {
	var t time.Time
	r.ParseForm()
	if ta := r.Form.Get("t"); len(ta) > 0 {
		t64, err := strconv.ParseInt(ta, 16, 64)
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintln(w, "wrong t:", ta)
			return
		}

		t = time.Unix(0, t64)
	}

	var j jsonWatchHistory
	c := f.WatchHistory(t)
	select {
	case e := <-c:
		switch e := e.(type) {
		case bool:
			if e {
				j.Info = "restarted"
			} else {
				j.Info = "unknown stopped"
			}
		case []*life.HistoryEvent:
			j.History, _ = formatHistoryEventDataList(e, profileIP, f)
		default:
			w.WriteHeader(500)
			fmt.Fprintln(w, "internal error: unknown chan type of watching history")
			return
		}
	case <-time.NewTimer(30 * time.Second).C:
		f.StopWatchHistory(c)
		j.Info = "timeout"
	}

	bytes, err := json.Marshal(j)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintln(w, err)
	} else {
		w.Write(bytes)
	}
}

func mapIncoming(in *life.Incoming) map[string]interface{} {
	m := make(map[string]interface{})
	m["id"] = in.ID()
	m["t"] = fmt.Sprintf("%016x", in.T().UnixNano())
	m["key"] = in.Key()
	m["start"] = strconv.FormatInt(in.Start().UnixNano()/1000000, 10)
	if !in.End().IsZero() {
		m["end"] = strconv.FormatInt(in.End().UnixNano()/1000000, 10)
	}

	return m
}

func (p *Proxy) watchIncoming(w http.ResponseWriter, r *http.Request, profileIP string, f *life.Life) {
	var t time.Time
	r.ParseForm()
	if ta := r.Form.Get("t"); len(ta) > 0 {
		t64, err := strconv.ParseInt(ta, 16, 64)
		if err != nil {
			w.WriteHeader(404)
			fmt.Fprintln(w, err)
			return
		}

		t = time.Unix(0, t64)
	}

	v := <-f.WatchIncoming(t)
	j := make([]interface{}, 0)
	switch v := v.(type) {
	case []*life.Incoming:
		for _, in := range v {
			j = append(j, mapIncoming(in))
		}
	case *life.Incoming:
		j = append(j, mapIncoming(v))
	}

	bytes, err := json.Marshal(j)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintln(w, err)
	} else {
		w.Write(bytes)
	}
}
