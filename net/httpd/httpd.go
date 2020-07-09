package httpd

import (
	"fmt"
	"net"
	"net/http"
)

type HttpHandler interface {
	GetDescription() string
	GetHandlePath() string
	OnRequest(w http.ResponseWriter, r *http.Request)
}

type tls struct {
	certFile string
	keyFile  string
}

type Http struct {
	serverAddress string
	tls           *tls
	times         int
	handlers      map[string]HttpHandler
}

func NewHttp() *Http {
	h := new(Http)
	return h
}

func NewHttps(certFile, keyFile string) *Http {
	h := new(Http)
	h.tls = &tls{certFile, keyFile}

	return h
}

func (h *Http) Scheme() string {
	if h.tls == nil {
		return "http"
	} else {
		return "https"
	}
}

func (h *Http) Init(server string) {
	h.serverAddress = server
	h.handlers = make(map[string]HttpHandler)
}

func (h *Http) RegisterHandler(handler HttpHandler) {
	h.handlers[handler.GetHandlePath()] = handler
}

func (h *Http) GetServerAddress() string {
	addr := h.serverAddress
	if len(addr) == 0 {
		addr = "localhost:4000"
	} else if addr[0] == ':' {
		addr = "localhost" + addr
	}

	return addr
}

func (h *Http) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request) {
	h.times++

	if len(r.URL.Scheme) == 0 && h.tls != nil {
		r.URL.Scheme = "https"
	}

	path := r.URL.Path
	handler, ok := h.handlers[path]
	if ok {
		handler.OnRequest(w, r)
	} else if handler, ok := h.handlers["/"]; ok {
		handler.OnRequest(w, r)
	} else {
		fmt.Fprintln(w, fmt.Sprintf("http object: %p, times: %d", h, h.times))
		fmt.Fprintln(w, "path: "+path)
		fmt.Fprintln(w, "method: "+r.Method)
		fmt.Fprintln(w, "url: "+r.RequestURI)
	}
}

func (h *Http) Run(e func(err error)) {
	var err error = nil
	defer func() { e(err) }()

	if h.tls == nil {
		err = http.ListenAndServe(h.serverAddress, h)
	} else {
		err = http.ListenAndServeTLS(h.serverAddress, h.tls.certFile, h.tls.keyFile, h)
	}
}

func RemoteHost(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	} else {
		return host
	}
}
