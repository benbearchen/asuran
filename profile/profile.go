package profile

import (
	"github.com/benbearchen/asuran/policy"

	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type urlAction struct {
	UrlPattern string
	pattern    *UrlPattern
	p          *policy.UrlPolicy
}

type UrlOperator interface {
	Action(ip, url string) *policy.UrlPolicy
}

type DomainAction struct {
	Domain  string
	pattern *DomainPattern
	p       *policy.DomainPolicy
}

type DomainOperator interface {
	Action(ip, domain string) *policy.DomainPolicy
}

func (d *DomainAction) TargetString() string {
	t := d.p.TargetString()
	if t == "" {
		t = "真实地址"
	}

	return t
}

type Store struct {
	ID      string
	Content []byte
}

type ProxyHostOperator interface {
	New(port int)
}

type Profile struct {
	Name       string
	Ip         string
	Owner      string
	Operators  map[string]bool
	Urls       map[string]*urlAction
	UrlDefault *policy.UrlPolicy
	Domains    map[string]*DomainAction
	storeID    int
	stores     map[string]*Store
	saver      *ProfileRootDir
	notSet     bool

	proxyOp ProxyHostOperator

	accessCode string

	lock sync.RWMutex
}

func NewProfile(name, ip, owner string, saver *ProfileRootDir) *Profile {
	p := new(Profile)
	p.Name = name
	p.Ip = ip
	p.Owner = owner
	p.Operators = make(map[string]bool)
	p.Urls = make(map[string]*urlAction)
	p.UrlDefault = policy.NewDefaultUrlPolicy()
	p.Domains = make(map[string]*DomainAction)
	p.storeID = 1
	p.stores = make(map[string]*Store)
	p.saver = saver
	p.notSet = true
	p.accessCode = makeRandomAccessCode()
	return p
}

func (p *Profile) SetUrlPolicy(s *policy.UrlPolicy, context *policy.PluginContext, op policy.PluginOperator) {
	p.lock.Lock()
	defer p.lock.Unlock()

	urlPattern := s.Target()
	if urlPattern == "" {
		p.pluginUpdate(p.UrlDefault, s, context, op)
		p.UrlDefault.Update(s)
		return
	}

	all := urlPattern == "all"

	if s.Delete() {
		if all {
			for up, u := range p.Urls {
				p.pluginRemove(u, context, op)
				delete(p.Urls, up)
			}
		} else {
			u, ok := p.Urls[urlPattern]
			if ok {
				p.pluginRemove(u, context, op)
				delete(p.Urls, urlPattern)
			}
		}

		return
	}

	if all {
		for _, u := range p.Urls {
			p.pluginUpdate(u.p, s, context, op)
			u.p.Update(s)
		}
	} else if u, ok := p.Urls[urlPattern]; ok {
		p.pluginUpdate(u.p, s, context, op)
		u.p.Update(s)
	} else {
		s.Def(p.UrlDefault)
		u := &urlAction{urlPattern, NewUrlPattern(urlPattern), s}
		p.Urls[urlPattern] = u
		if p.proxyOp != nil && u.pattern != nil && len(u.pattern.port) > 0 {
			if port, err := strconv.Atoi(u.pattern.port); err == nil {
				p.proxyOp.New(port)
			}
		}

		host := getHostOfUrlPattern(urlPattern)
		if len(host) != 0 {
			p.proxyDomainIfNotExists(host)
		}

		p.pluginUpdate(u.p, s, context, op)
	}
}

func (p *Profile) UrlAction(url string) *policy.UrlPolicy {
	p.lock.RLock()
	defer p.lock.RUnlock()

	u := p.MatchUrl(url)
	if u != nil {
		return u.p
	}

	return p.UrlDefault
}

func (p *Profile) MatchUrl(url string) *urlAction {
	us := parseUrlSection(url)
	var high uint32 = 0
	var highUrl *urlAction = nil
	for _, u := range p.Urls {
		score := u.pattern.MatchScore(us)
		if score > high {
			high = score
			highUrl = u
		}
	}

	return highUrl
}

func (p *Profile) DeleteAllUrl() {
	p.lock.Lock()
	defer p.lock.Unlock()

	for u, _ := range p.Urls {
		delete(p.Urls, u)
	}
}

func (p *Profile) SetDomainPolicy(s *policy.DomainPolicy) {
	p.lock.Lock()
	defer p.lock.Unlock()

	domain := s.Domain()
	all := domain == "all"
	if s.Delete() {
		if all {
			for d, _ := range p.Domains {
				delete(p.Domains, d)
			}
		} else {
			delete(p.Domains, domain)
		}

		return
	}

	if all {
		for _, d := range p.Domains {
			d.p.Update(s)
		}
	} else if d, ok := p.Domains[domain]; ok {
		d.p.Update(s)
	} else {
		p.Domains[domain] = &DomainAction{domain, NewDomainPattern(domain), s}
	}
}

func (p *Profile) Domain(domain string) *policy.DomainPolicy {
	p.lock.RLock()
	defer p.lock.RUnlock()

	d, ok := p.Domains[domain]
	if ok {
		return d.p
	}

	for _, d := range p.Domains {
		if d.pattern != nil && d.pattern.Match(domain) {
			return d.p
		}
	}

	return nil
}

func (p *Profile) ProxyDomainIfNotExists(domain string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.proxyDomainIfNotExists(domain)
}

func (p *Profile) proxyDomainIfNotExists(domain string) {
	if len(domain) == 0 {
		return
	}

	d, ok := p.Domains[domain]
	if ok {
		if d.p.Action() == nil {
			d.p.SetProxy()
		}
		return
	}

	// TODO: 更好地构造 DomainPolicy？
	dp, err := policy.Factory("domain proxy " + domain)
	if err != nil {
		fmt.Println("set domain", domain, "failed,", err)
		return
	}

	p.Domains[domain] = &DomainAction{domain, NewDomainPattern(domain), dp.(*policy.DomainPolicy)}
}

func (p *Profile) DeleteDomain(domain string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.Domains, domain)
}

func (p *Profile) DeleteAllDomain() {
	p.lock.Lock()
	defer p.lock.Unlock()

	for d, _ := range p.Domains {
		delete(p.Domains, d)
	}
}

func (p *Profile) Delete(urlPattern string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.Urls, urlPattern)
}

func (p *Profile) Store(id string, content []byte) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.stores[id] = &Store{id, content}
}

func (p *Profile) StoreID(content []byte) string {
	p.lock.Lock()
	defer p.lock.Unlock()

	for {
		id := "s" + strconv.Itoa(p.storeID)
		p.storeID++
		if _, ok := p.stores[id]; ok {
			continue
		}

		p.stores[id] = &Store{id, content}
		return id
	}
}

func (p *Profile) Restore(id string) []byte {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if b, ok := p.stores[id]; ok {
		return b.Content
	} else {
		return nil
	}
}

func (p *Profile) DeleteStore(id string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.stores, id)
}

func (p *Profile) DeleteAllStore() {
	p.lock.Lock()
	defer p.lock.Unlock()

	for id, _ := range p.stores {
		delete(p.stores, id)
	}
}

func (p *Profile) ListStoreIDs() []string {
	p.lock.RLock()
	defer p.lock.RUnlock()

	s := make([]string, 0, len(p.stores))
	for _, v := range p.stores {
		s = append(s, v.ID)
	}

	return s
}

func (p *Profile) ListStored() []*Store {
	p.lock.RLock()
	defer p.lock.RUnlock()

	s := make([]*Store, 0, len(p.stores))
	for _, v := range p.stores {
		s = append(s, v)
	}

	return s
}

func (p *Profile) AddOperator(ip string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if len(ip) > 0 {
		p.Operators[ip] = true
	}
}

func (p *Profile) RemoveOperator(ip string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if len(ip) > 0 {
		delete(p.Operators, ip)
	}
}

func (p *Profile) CanOperate(ip string) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if p.Owner == ip || p.Ip == ip {
		return true
	}

	_, ok := p.Operators[ip]
	if ok {
		return true
	}

	return false
}

func (p *Profile) CloneNew(newName, newIp, newOwner string) *Profile {
	p.lock.RLock()
	defer p.lock.RUnlock()

	n := NewProfile(newName, newIp, newOwner, p.saver)
	n.proxyOp = p.proxyOp
	for u, url := range p.Urls {
		c := *url
		n.Urls[u] = &c
	}

	for d, domain := range p.Domains {
		c := *domain
		n.Domains[d] = &c
	}

	for s, store := range p.stores {
		c := *store
		n.stores[s] = &c
	}

	return n
}

func (p *Profile) Clear() {
	p.DeleteAllUrl()
	p.DeleteAllDomain()
	p.storeID = 1
	p.DeleteAllStore()
}

func (p *Profile) AccessCode() string {
	return p.accessCode
}

func (p *Profile) CheckAccessCode(accessCode string) bool {
	return strings.ToLower(accessCode) == p.accessCode
}

func (p *Profile) Load() string {
	c, _ := p.saver.Load(p.Ip)
	return string(c)
}

func (p *Profile) Save() {
	p.saver.Save(p.Ip, p.ExportCommand())
	p.notSet = false
}

func (p *Profile) pluginUpdate(u, up *policy.UrlPolicy, context *policy.PluginContext, op policy.PluginOperator) {
	if op == nil {
		return
	}

	old := u.Plugin()
	n := up.Plugin()
	if old == nil && n == nil {
		return
	}

	removeOld := false
	if up.Set() {
		if old != nil {
			if n == nil || n.Name() != old.Name() {
				removeOld = true
			}
		}
	} else {
		if old != nil && n != nil && n.Name() != old.Name() {
			removeOld = true
		}
	}

	if removeOld {
		op.Remove(context, old.Name())
	}

	if n != nil {
		op.Update(context, n.Name(), n)
	}
}

func (p *Profile) pluginRemove(u *urlAction, context *policy.PluginContext, op policy.PluginOperator) {
	if op == nil {
		return
	}

	old := u.p.Plugin()
	if old == nil {
		return
	}

	op.Remove(context, old.Name())
}

func getHostOfUrlPattern(urlPattern string) string {
	p := strings.Index(urlPattern, "://")
	if p >= 0 {
		urlPattern = urlPattern[p+3:]
	}

	p = strings.Index(urlPattern, "/")
	if p == 0 {
		return ""
	} else if p < 0 {
		p = len(urlPattern)
	}

	server := urlPattern[0:p]
	host, _, err := net.SplitHostPort(server)
	if err != nil {
		return server
	} else {
		return host
	}
}

var (
	randGen = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func makeRandomAccessCode() string {
	c := randGen.Uint32()
	return fmt.Sprintf("%8x", c)
}
