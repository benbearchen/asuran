package api

import (
	"github.com/benbearchen/asuran/policy"

	"fmt"
	"net/http"
	"sync"
)

type Context struct {
	ProfileIP string // may be empty.
	Policy    *policy.PluginPolicy
	Log       func(statusCode int, postBody, content []byte, err error)
}

type API interface {
	Name() string
	Intro() string
	Call(context *Context, targetURI string, w http.ResponseWriter, r *http.Request)
	Reset(context *Context)
}

var (
	plugins map[string]API = make(map[string]API)
	lock    sync.RWMutex
)

func Register(plugin API) {
	lock.Lock()
	defer lock.Unlock()

	plugins[plugin.Name()] = plugin
}

type apiHandler struct {
	name    string
	intro   string
	handler func(context *Context, targetURI string, w http.ResponseWriter, r *http.Request)
}

func RegisterHandler(name, intro string, handler func(context *Context, targetURI string, w http.ResponseWriter, r *http.Request)) {
	Register(&apiHandler{name, intro, handler})
}

func All() []string {
	lock.RLock()
	defer lock.RUnlock()

	names := make([]string, 0, len(plugins))

	for name := range plugins {
		names = append(names, name)
	}

	return names
}

func Call(context *Context, name string, targetURI string, w http.ResponseWriter, r *http.Request) {
	lock.RLock()
	defer lock.RUnlock()

	plugin, ok := plugins[name]
	if ok {
		plugin.Call(context, targetURI, w, r)
	} else {
		w.WriteHeader(500)
		fmt.Fprintln(w, "miss plugin: "+name)
	}
}

func Reset(context *Context) {
	lock.RLock()
	defer lock.RUnlock()

	for _, p := range plugins {
		p.Reset(context)
	}
}

func Intro(name string) string {
	lock.RLock()
	defer lock.RUnlock()

	plugin, ok := plugins[name]
	if ok {
		return plugin.Intro()
	} else {
		return ""
	}
}

func (h *apiHandler) Name() string {
	return h.name
}

func (h *apiHandler) Intro() string {
	return h.intro
}

func (h *apiHandler) Call(context *Context, targetURI string, w http.ResponseWriter, r *http.Request) {
	h.handler(context, targetURI, w, r)
}

func (h *apiHandler) Reset(context *Context) {
}
