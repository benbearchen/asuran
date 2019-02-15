package api

import (
	"github.com/benbearchen/asuran/policy"

	"fmt"
	"net/http"
	"sync"
)

type API interface {
	Name() string
	Intro() string

	Update(context *policy.PluginContext, p *policy.PluginPolicy)
	Remove(context *policy.PluginContext)
	Reset(context *policy.PluginContext)

	Call(context *policy.PluginContext, w http.ResponseWriter, r *http.Request)
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
	name     string
	intro    string
	policies map[string]map[string]*policy.PluginPolicy // map[ip][url]*policy
	handler  func(context *policy.PluginContext, p *policy.PluginPolicy, w http.ResponseWriter, r *http.Request)
}

func RegisterHandler(name, intro string, handler func(context *policy.PluginContext, p *policy.PluginPolicy, w http.ResponseWriter, r *http.Request)) {
	policies := make(map[string]map[string]*policy.PluginPolicy)
	Register(&apiHandler{name, intro, policies, handler})
}

func Has(name string) bool {
	lock.RLock()
	defer lock.RUnlock()

	_, ok := plugins[name]
	return ok
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

func Call(context *policy.PluginContext, name string, w http.ResponseWriter, r *http.Request) {
	lock.RLock()
	defer lock.RUnlock()

	plugin, ok := plugins[name]
	if ok {
		plugin.Call(context, w, r)
	} else {
		w.WriteHeader(500)
		fmt.Fprintln(w, "call miss plugin: "+name)
	}
}

func Update(context *policy.PluginContext, name string, p *policy.PluginPolicy) {
	lock.RLock()
	defer lock.RUnlock()

	plugin, ok := plugins[name]
	if ok {
		plugin.Update(context, p)
	} else {
		fmt.Println("update miss plugin: " + name)
	}
}

func Remove(context *policy.PluginContext, name string) {
	lock.RLock()
	defer lock.RUnlock()

	plugin, ok := plugins[name]
	if ok {
		plugin.Remove(context)
	} else {
		fmt.Println("remove miss plugin: " + name)
	}
}

func Reset(context *policy.PluginContext, name string) {
	lock.RLock()
	defer lock.RUnlock()

	if len(name) == 0 {
		for _, p := range plugins {
			p.Reset(context)
		}
	} else {
		plugin, ok := plugins[name]
		if ok {
			plugin.Remove(context)
		} else {
			fmt.Println("reset miss plugin: " + name)
		}
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

func (h *apiHandler) Call(context *policy.PluginContext, w http.ResponseWriter, r *http.Request) {
	s, ok := h.policies[context.ProfileIP]
	if !ok {
		w.WriteHeader(500)
		fmt.Fprintln(w, "call plugin miss session: "+context.ProfileIP)
		return
	}

	p, ok := s[context.TargetURL]
	if !ok {
		w.WriteHeader(500)
		fmt.Fprintln(w, "call plugin miss target: "+context.TargetURL+" of session "+context.ProfileIP)
		return
	}

	h.handler(context, p, w, r)
}

func (h *apiHandler) Update(context *policy.PluginContext, p *policy.PluginPolicy) {
	if len(context.ProfileIP) == 0 || len(context.TargetURL) == 0 {
		return
	}

	s, ok := h.policies[context.ProfileIP]
	if !ok {
		return
	}

	s[context.TargetURL] = p
}

func (h *apiHandler) Remove(context *policy.PluginContext) {
	if len(context.ProfileIP) == 0 {
		h.policies = make(map[string]map[string]*policy.PluginPolicy)
		return
	}

	if len(context.TargetURL) == 0 {
		delete(h.policies, context.ProfileIP)
		return
	}

	s, ok := h.policies[context.ProfileIP]
	if !ok {
		return
	}

	delete(s, context.TargetURL)
}

func (h *apiHandler) Reset(context *policy.PluginContext) {
}
