package api

import (
	"sync"
)

type API interface {
	Name() string
	Intro() string
	Link() string
	ShowLink() bool

	Init()
}

var (
	tuns map[string]API = make(map[string]API)
	lock sync.Mutex
)

func Register(tun API) {
	lock.Lock()
	defer lock.Unlock()

	tuns[tun.Name()] = tun
	go tun.Init()
}

func RegisterLinker(name, intro, link string, show bool) {
	s := MakeSimpleLinker(name, intro, link, show)
	Register(s)
}

func List() []API {
	lock.Lock()
	defer lock.Unlock()

	apis := make([]API, 0, len(tuns))
	for _, tun := range tuns {
		apis = append(apis, tun)
	}

	return apis
}

func Get(name string) API {
	lock.Lock()
	defer lock.Unlock()

	tun, _ := tuns[name]
	return tun
}

type Simple struct {
	name  string
	intro string
	link  string
	show  bool
}

func MakeSimpleLinker(name, intro, link string, show bool) *Simple {
	s := new(Simple)
	s.name = name
	s.intro = intro
	s.link = link
	s.show = show

	return s
}

func (s *Simple) Name() string {
	return s.name
}

func (s *Simple) Intro() string {
	return s.intro
}

func (s *Simple) Link() string {
	return s.link
}

func (s *Simple) ShowLink() bool {
	return s.show
}

func (s *Simple) Init() {
	// nothing to do
}
