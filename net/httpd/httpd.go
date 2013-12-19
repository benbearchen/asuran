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

type Http struct {
	serverAddress string
	times         int
	handlers      map[string]HttpHandler
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

func (h *Http) Run() {
	http.ListenAndServe(h.serverAddress, h)
}

func RemoteHost(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	} else {
		return host
	}
}
