package proxy

import (
	"github.com/benbearchen/asuran/profile"
	"github.com/benbearchen/asuran/web/proxy/cache"
	"github.com/benbearchen/asuran/web/proxy/life"
	"github.com/benbearchen/asuran/web/proxy/pack"
	"github.com/benbearchen/asuran/web/proxy/plugin/api"
	tunnel "github.com/benbearchen/asuran/web/proxy/tunnel/api"

	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type htmlUsage struct {
	IP       string
	Host     string
	Domain   string
	UsingDNS bool
}

func (p *Proxy) WriteUsage(w io.Writer) {
	t, err := template.ParseFiles("template/usage.tmpl")

	u := htmlUsage{}
	u.IP = p.serveIP
	u.Host = p.mainHost
	u.Domain = p.domain
	u.UsingDNS = !p.disableDNS

	err = t.Execute(w, u)
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

type htmlInitDevice struct {
	ProxyIP  string
	ClientIP string
}

func (p *Proxy) WriteInitDevice(w io.Writer, ip string) {
	t, err := template.ParseFiles("template/domain.tmpl")

	d := htmlInitDevice{}
	d.ProxyIP = p.serveIP
	d.ClientIP = ip

	err = t.Execute(w, d)
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

type indexData struct {
	Version    string
	ServeIP    string
	MainHost   string
	ProxyHost  string
	UsingDNS   bool
	Client     string
	AccessCode string
}

func (p *Proxy) index(w http.ResponseWriter, ver, clientIP, accessCode string) {
	t, err := template.ParseFiles("template/index.tmpl")
	err = t.Execute(w, indexData{ver, p.serveIP, p.mainHost, p.proxyAddr, !p.disableDNS, clientIP, accessCode})
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

func (p *Proxy) features(w http.ResponseWriter, ver string) {
	t, err := template.ParseFiles("template/features.tmpl")
	err = t.Execute(w, indexData{ver, p.serveIP, p.mainHost, p.proxyAddr, !p.disableDNS, "", ""})
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

func (p *Proxy) res(w http.ResponseWriter, r *http.Request, path string) {
	f := filepath.Clean(path)
	h := filepath.Clean("/res/")
	if !strings.HasPrefix(f, h) || len(f) <= len(h) {
		w.WriteHeader(403)
		return
	}

	dir, err := os.Getwd()
	if err != nil {
		w.WriteHeader(404)
		return
	}

	f = filepath.Join(dir, "template", f)
	http.ServeFile(w, r, f)
}

func (p *Proxy) dir(w http.ResponseWriter, r *http.Request, path string) {
	entry := "/dir/"
	if !strings.HasPrefix(path, entry) {
		w.WriteHeader(403)
		return
	}

	path = filepath.Clean(path[len(entry):])
	d, n := filepath.Split(path)
	dir := p.lookupMapDir(filepath.Clean(d))
	if len(dir) <= 0 {
		w.WriteHeader(404)
		return
	}

	n = filepath.Join(dir, n)
	http.ServeFile(w, r, n)
}

type urlHistoryData struct {
	Even         bool
	ID           string
	Time         string
	Client       string
	Method       string
	ResponseCode string
	RecvBytes    string
}

type urlHistoryListData struct {
	Client    string
	Url       string
	Histories []urlHistoryData
}

func formatUrlHistoryDataList(histories []*cache.UrlHistory, client string) []urlHistoryData {
	d := make([]urlHistoryData, 0, len(histories))
	even := true
	for _, h := range histories {
		even = !even

		responseCode := ""
		if h.ResponseCode >= 0 {
			responseCode = strconv.Itoa(h.ResponseCode)
		}

		recvBytes := "出错"
		if h.Bytes != nil {
			recvBytes = strconv.Itoa(len(h.Bytes))
		}

		d = append(d, urlHistoryData{even, strconv.FormatUint(uint64(h.ID), 10), h.Time.Format("2006-01-02 15:04:05"), client, h.Method, responseCode, recvBytes})
	}

	return d
}

func (p *Proxy) writeUrlHistoryList(w http.ResponseWriter, profileIP, url string, histories []*cache.UrlHistory) {
	t, err := template.ParseFiles("template/history.urllist.tmpl")
	err = t.Execute(w, urlHistoryListData{profileIP, url, formatUrlHistoryDataList(histories, profileIP)})
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

type opData struct {
	Name   string `json:"name"`
	Act    string `json:"act"`
	Arg    string `json:"arg"`
	Client string `json:"client"`
}

type historyEventData struct {
	Even        bool
	T           string   `json:"t"`
	Time        string   `json:"time"`
	Domain      string   `json:"domain"`
	DomainIP    string   `json:"domainIP"`
	URL         string   `json:"url"`
	URLID       string   `json:"urlID"`
	URLBody     string   `json:"urlBody"`
	HttpStatus  string   `json:"info"`
	EventString string   `json:"log"`
	OPs         []opData `json:"ops"`
	Client      string   `json:"client"`
}

type historyData struct {
	Client string
	LastT  string
	Events []historyEventData
}

func formatHistoryEventDataList(events []*life.HistoryEvent, client string, f *life.Life) ([]historyEventData, string) {
	list := make([]historyEventData, 0, len(events))
	even := true
	lastT := ""
	for _, e := range events {
		d := historyEventData{}
		d.OPs = make([]opData, 0)
		d.Client = client

		even = !even
		d.Even = even
		d.Time = e.Time.Format("2006-01-02 15:04:05")
		d.T = fmt.Sprintf("%016x", e.Time.UnixNano())
		lastT = d.T

		s := strings.Split(e.String, " ")
		if len(s) >= 3 && s[0] == "domain" {
			domain := s[2]
			d.Domain = "域名 " + s[1] + " " + domain
			d.OPs = append(d.OPs, opData{"代理域名", "domain/redirect", domain, client})
			if len(s) >= 4 {
				d.DomainIP = s[3]
			}
		} else if len(s) >= 3 && s[0] == "proxy" {
			url := s[1]
			d.URL = url

			if s[2] == "redirect" {
				d.HttpStatus = "重定向"
				if len(s) >= 4 {
					d.URL += " => " + s[3]
				}
			} else if id, err := strconv.ParseInt(s[2], 10, 32); err == nil {
				d.URLID = s[2]
				h := f.LookHistoryByID(uint32(id))
				if h != nil {
					status := h.Method
					if h.ResponseCode >= 0 {
						status += " " + strconv.Itoa(h.ResponseCode)
					} else {
						status += " 出错"
					}

					d.HttpStatus = status
				}

				if len(s) >= 4 {
					d.URL += " " + s[3]
				}

				d.OPs = append(d.OPs, opData{"缓存", "url/store", d.URLID, client})
			} else {
				d.HttpStatus = s[2]
				if len(s) >= 4 {
					d.HttpStatus += " " + s[3]
				}
			}

			if strings.HasPrefix(url, "http://") {
				d.URLBody = url[7:]
			} else {
				d.URLBody = url
			}
		} else if len(s) >= 2 && s[0] == "cache" {
			d.HttpStatus = "命中缓存"
			url := s[1]
			d.URL = url
			if len(s) > 2 {
				d.URL += " " + s[2]
			}

			if strings.HasPrefix(url, "http://") {
				d.URLBody = url[7:]
			} else {
				d.URLBody = url
			}
		} else if len(s) >= 4 && s[0] == "plugin" {
			d.HttpStatus = "插件 " + s[3]

			url := s[1]
			d.URL = url

			if id, err := strconv.ParseInt(s[2], 10, 32); err == nil {
				d.URLID = s[2]
				h := f.LookHistoryByID(uint32(id))
				if h != nil {
					status := h.Method
					if h.ResponseCode >= 0 {
						status += " " + strconv.Itoa(h.ResponseCode)
					} else {
						status += " 出错"
					}

					d.HttpStatus += ", " + status
				}
			}

			if strings.HasPrefix(url, "http://") {
				d.URLBody = url[7:]
			} else {
				d.URLBody = url
			}
		} else {
			d.EventString = e.String
		}

		list = append(list, d)
	}

	return list, lastT
}

func (p *Proxy) writeHistory(w http.ResponseWriter, profileIP string, f *life.Life) {
	t, err := template.ParseFiles("template/history.tmpl")
	//list, lastT := formatHistoryEventDataList(f.HistoryEvents(), profileIP, f)
	err = t.Execute(w, historyData{profileIP, "", []historyEventData{}})
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

type dnsHistoryEventData struct {
	Even     bool
	Profile  bool
	Time     string
	Domain   string
	DomainIP string
	Client   string
}

type dnsHistoryData struct {
	Target string
	Events []dnsHistoryEventData
}

func formatDNSHistoryEventDataList(events []*life.HistoryEvent, f *life.Life, targetIP string) []dnsHistoryEventData {
	list := make([]dnsHistoryEventData, 0, len(events))
	even := true
	for _, e := range events {
		d := dnsHistoryEventData{}

		even = !even
		d.Even = even
		d.Time = e.Time.Format("2006-01-02 15:04:05")

		s := strings.Split(e.String, " ")
		if len(s) >= 3 {
			client := s[0]
			if len(targetIP) > 0 {
				if targetIP != client {
					continue
				} else {
					d.Profile = true
				}
			}

			d.Client = client

			domain := s[2]
			d.Domain = "域名 " + s[1] + " " + domain
			if len(s) >= 4 {
				d.DomainIP = s[3]
			}
		} else {
			continue
		}

		list = append(list, d)
	}

	return list
}

func (p *Proxy) writeDNSHistory(w http.ResponseWriter, f *life.Life, targetIP string) {
	t, err := template.ParseFiles("template/dnshistory.tmpl")
	list := formatDNSHistoryEventDataList(f.HistoryEvents(), f, targetIP)
	targetInfo := ""
	if len(targetIP) == 0 {
		targetInfo = "DNS 服务"
	} else {
		targetInfo = targetIP + " DNS"
	}

	err = t.Execute(w, dnsHistoryData{targetInfo, list})
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

type deviceData struct {
	Even       bool
	Name       string
	IP         string
	Owner      string
	InitTime   string
	VisitTime  string
	ActiveTime string
}

type devicesListData struct {
	Devices []deviceData
}

func formatDevicesListData(profiles []*profile.Profile, v *life.IPLives) devicesListData {
	devices := make([]deviceData, 0)
	if len(profiles) > 0 {
		index := make(map[string]*profile.Profile)
		ips := make([]string, 0, len(profiles))
		for _, p := range profiles {
			if p.Ip == "localhost" {
				continue
			}

			index[p.Ip] = p
			ips = append(ips, p.Ip)
		}

		even := true
		sort.Strings(ips)
		for _, ip := range ips {
			p := index[ip]

			it := ""
			vt := ""
			at := ""
			f := v.OpenExists(p.Ip)
			if f != nil {
				format := func(t time.Time) string {
					if t.IsZero() {
						return ""
					} else {
						return t.Format("2006-01-02 15:04:05")
					}
				}

				it = format(f.CreateTime)
				vt = format(f.VisitTime)
				at = format(f.ActiveTime)
			}

			even = !even

			devices = append(devices, deviceData{even, p.Name, p.Ip, p.Owner, it, vt, at})
		}
	}

	return devicesListData{devices}
}

func (p *Proxy) devices(w http.ResponseWriter) {
	t, err := template.ParseFiles("template/devices.tmpl")
	profiles := make([]*profile.Profile, 0)
	if p.profileOp != nil {
		profiles = p.profileOp.All()
	}

	err = t.Execute(w, formatDevicesListData(profiles, p.lives))
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

func (p *Proxy) urlEncoded(w http.ResponseWriter) {
	t, err := template.ParseFiles("template/urlencoded.tmpl")
	err = t.Execute(w, nil)
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

type storeData struct {
	Even           bool
	Client         string
	ID             string
	EncodedContent string
}

type storeListData struct {
	Client   string
	Contents []storeData
}

func formatStoreListData(saved []*profile.Store, profileIP string) []storeData {
	s := make([]storeData, 0, len(saved))
	even := true
	for _, v := range saved {
		even = !even
		s = append(s, storeData{even, profileIP, v.ID, url.QueryEscape(string(v.Content))})
	}

	return s
}

func (p *Proxy) writeStores(w http.ResponseWriter, profileIP string, prof *profile.Profile) {
	t, err := template.ParseFiles("template/stores.tmpl")
	list := formatStoreListData(prof.ListStored(), profileIP)
	err = t.Execute(w, storeListData{profileIP, list})
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

type editStoreData struct {
	Client         string
	ID             string
	EncodedContent string
	View           bool
}

func formatEditStoreData(profileIP string, prof *profile.Profile, id string, view bool) editStoreData {
	encodedContent := ""
	if len(id) > 0 {
		c := prof.Restore(id)
		if len(c) > 0 {
			encodedContent = strings.Replace(url.QueryEscape(string(c)), "+", "%20", -1)
		}
	}

	return editStoreData{profileIP, id, encodedContent, view}
}

func (p *Proxy) writeEditStore(w http.ResponseWriter, profileIP string, prof *profile.Profile, id string, view bool) {
	t, err := template.ParseFiles("template/store-edit.tmpl")
	err = t.Execute(w, formatEditStoreData(profileIP, prof, id, view))
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

type storeResultData struct {
	IP  string
	URL string
	ID  string
	SID string
}

func (p *Proxy) writeStoreResult(w http.ResponseWriter, profileIP, url, id, sid string) {
	t, err := template.ParseFiles("template/store-result.tmpl")
	err = t.Execute(w, storeResultData{profileIP, url, id, sid})
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

type packData struct {
	Even    bool
	Name    string
	Author  string
	Comment string
}

type packsData struct {
	Packs []packData
}

func formatPacksData(packs *pack.Dir) packsData {
	names := packs.ListNames()
	datas := make([]packData, 0, len(names))
	for i, name := range names {
		pack := packs.GetPack(name)
		even := i%2 == 1
		data := packData{even, pack.Name(), pack.Author(), pack.Comment()}
		datas = append(datas, data)
	}

	return packsData{datas}
}

func (p *Proxy) writePacks(w http.ResponseWriter) {
	t, err := template.ParseFiles("template/packs-list.tmpl")
	err = t.Execute(w, formatPacksData(p.packs))
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

type pluginData struct {
	Even  bool
	Name  string
	Intro string
}

type pluginsData struct {
	Plugins []pluginData
}

func formatPluginsData() pluginsData {
	names := api.All()
	datas := make([]pluginData, 0, len(names))
	for i, name := range names {
		intro := api.Intro(name)
		datas = append(datas, pluginData{i%2 == 1, name, intro})
	}

	return pluginsData{datas}
}

func (p *Proxy) writePlugins(w http.ResponseWriter) {
	t, err := template.ParseFiles("template/plugins-list.tmpl")
	err = t.Execute(w, formatPluginsData())
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}

type tunnelData struct {
	Even  bool
	Name  string
	Intro string
	Entry string
	Show  bool
	Link  string
}

type tunnelsData struct {
	Tunnels []tunnelData
}

func formatTunnelsData(prefix string) tunnelsData {
	tuns := tunnel.List()
	sort.Slice(tuns, func(i, j int) bool {
		return tuns[i].Name() < tuns[j].Name()
	})

	datas := make([]tunnelData, len(tuns))
	for i, tun := range tuns {
		datas[i].Even = (i%2 == 1)
		datas[i].Name = tun.Name()
		datas[i].Intro = tun.Intro()
		datas[i].Entry = prefix + "/" + tun.Name()
		datas[i].Show = tun.ShowLink()
		if tun.ShowLink() {
			datas[i].Link = tun.Link()
		}
	}

	return tunnelsData{datas}
}

func (p *Proxy) writeTunnels(w http.ResponseWriter) {
	t, err := template.ParseFiles("template/tunnels-list.tmpl")
	err = t.Execute(w, formatTunnelsData(p.tunnelPrefix()))
	if err != nil {
		fmt.Fprintln(w, "内部错误：", err)
	}
}
