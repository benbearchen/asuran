package profile

import (
	"github.com/benbearchen/asuran/util/cmd"

	"net"
	"strconv"
	"strings"
)

func CommandUsage() string {
	return `commands:
-------
# 以 # 开头的行为注释

delay [default] <duration> <url-pattern>
delay drop <duration> <url-pattern>

proxy [default] <url-pattern>
proxy cache <url-pattern>
proxy drop <responseCode> <url-pattern>

delete <url-pattern>

domain [default] <domain-name>
domain block <domain-name>
domain redirect <domain-name> [ip]

domain delete <domain-name>


compatible commands:
-------
ip <domain-name> 
# 等价于  domain redirect <domain-name> ip


-------
注：
* <> 尖括号表示参数一定要有，否则出错
* [] 中括号表示参数可有可无
* 下面注释以“**”开始的行，表示未实现功能
-------

delay mode:   只能处于一下模式之一种
    [default] 所有请求延时 duration；
              duration == 0 表示不延时，立即返回。
    drop      从 URL 第一次至 duration 时间内的请求一律丢弃，
              直到 duration 以后的请求正常返回。
              duration == 0 表示只丢弃第一次请求。
              ** “丢弃”的效果可能无法很好实现 **

proxy mode:   只能处于一下模式之一种
    [default] 每次重新代理请求。
    cache     缓存请求结果，下次请求起从缓存返回。
    drop <responseCode>
              丢弃请求，以 responseCode 回应。
              responseCode 可以是 404、502 等。

duration:
              时长，可选单位：ms, s, m, h。默认为 s
              例：90s 或 1.5m

url-pattern:
    [domain[:port]]/[path][?key=value]
              分域名[端口]、根路径与查询参数三种匹配。
              域名忽略则匹配所有域名。
              根路径可以匹配到目录或文件。
              查询参数匹配时忽略顺序，但列出参数必须全有。
              域名支持通配符“*”，如 *.com, *play.org
    all
              特殊地，all 可以操作所有已经配置的 URL。

domain mode:
    [default] 域名默认为正常通行，返回正常结果。
    block     屏蔽域名，不返回结果。
    redirect  把域名重定向到制定 IP。
              如果 IP 为空则重定向到代理服务器。

domain-name:
    ([^.]+.)+[^.]+
              域名，目前支持英文域名（中文域名未验证）。
    all
              特殊地，all 可以操作所有已经配置的域名。

ip:           IP 地址，比如 192.168.1.3。

-------
examples:

delay 5s g.cn/search

proxy default github.com/?cmd=1

proxy cache golang.org/doc/code.html

proxy drop 404 baidu.com/

domain g.cn

domain block baidu.com

domain redirect g.cn

domain delete g.cn
`
}

func (p *IpProfiles) Command(ip, command string) {
	profile := p.FindByIp(ip)
	if profile != nil {
		profile.Command(command)
	}
}

func (p *Profile) Command(command string) {
	commandLines := strings.Split(command, "\n")
	for _, line := range commandLines {
		line = strings.TrimSpace(line)
		if len(line) <= 0 || line[0] == '#' {
			continue
		}

		c, rest := cmd.TakeFirstArg(line)
		switch c {
		case "delay":
			p.CommandDelay(rest)
		case "proxy":
			p.CommandProxy(rest)
		case "delete":
			p.CommandDelete(rest)
		case "domain":
			p.CommandDomain(rest)
		default:
		}
	}
}

func (p *Profile) CommandDelay(content string) {
	c, rest := cmd.TakeFirstArg(content)
	switch c {
	case "default":
		commandDelayMode(p, c, rest)
	case "drop":
		commandDelayMode(p, c, rest)
	default:
		commandDelayMode(p, "default", content)
	}
}

func (p *Profile) CommandProxy(content string) {
	c, rest := cmd.TakeFirstArg(content)
	switch c {
	case "default":
		commandProxyMode(p, c, rest)
	case "cache":
		commandProxyMode(p, c, rest)
	case "drop":
		commandProxyMode(p, c, rest)
	default:
		commandProxyMode(p, "default", content)
	}
}

func (p *Profile) CommandDelete(content string) {
	if content == "all" {
		p.DeleteAllUrl()
		return
	}

	pattern := restToPattern(content)
	if len(pattern) > 0 {
		p.Delete(pattern)
	}
}

func (p *Profile) CommandDomain(content string) {
	c, rest := cmd.TakeFirstArg(content)
	switch c {
	case "default":
		commandDomainMode(p, c, rest)
	case "block":
		commandDomainMode(p, c, rest)
	case "redirect":
		commandDomainMode(p, c, rest)
	case "delete":
		commandDomainDelete(p, rest)
	default:
		commandDomainMode(p, "default", content)
	}
}

func restToPattern(content string) string {
	url, rest := cmd.TakeFirstArg(content)
	if len(rest) > 0 {
		return ""
	}

	if url == "all" {
		return url
	}

	if strings.HasPrefix(url, "http://") {
		url = url[len("http://"):]
	}

	q := strings.Index(url, "?")
	s := strings.Index(url, "/")
	if q >= 0 && s < 0 {
		url = url[0:q] + "/" + url[q:]
	} else if s < 0 {
		url = url + "/"
	}

	return url
}

func commandDelayMode(p *Profile, mode, args string) {
	var act DelayActType = DelayActDelayEach
	if mode == "drop" {
		act = DelayActDropUntil
	}

	duration, pattern, ok := delayTimeAndPattern(args)
	if ok {
		if act == DelayActDelayEach && duration == 0 {
			act = DelayActNone
		}

		if pattern == "all" {
			p.SetAllUrlDelay(act, duration)
		} else {
			p.SetUrlDelay(pattern, act, duration)
		}
	}
}

func delayTimeAndPattern(content string) (float32, string, bool) {
	d, p := cmd.TakeFirstArg(content)
	duration := parseDuration(d)
	pattern := restToPattern(p)
	ok := duration >= 0 && len(pattern) > 0
	return duration, pattern, ok
}

func parseDuration(d string) float32 {
	var times float64 = 1
	if strings.HasSuffix(d, "ms") {
		times = 0.001
	} else if strings.HasSuffix(d, "h") {
		d = d[:len(d)-1]
		times = 60 * 60
	} else if strings.HasSuffix(d, "m") {
		d = d[:len(d)-1]
		times = 60
	} else if strings.HasSuffix(d, "s") {
		d = d[:len(d)-1]
	}

	f, err := strconv.ParseFloat(d, 32)
	if err != nil {
		return -1
	} else {
		return float32(f * float64(times))
	}
}

func commandProxyMode(p *Profile, mode, args string) {
	var act UrlAct = UrlActNone
	if mode == "cache" {
		act = UrlActCache
	} else if mode == "drop" {
		act = UrlActDrop
	}

	dropResponseCode := 0
	if act == UrlActDrop {
		r, rest := cmd.TakeFirstArg(args)
		responseCode, err := strconv.Atoi(r)
		if err != nil {
			return
		} else {
			dropResponseCode = responseCode
			args = rest
		}
	}

	pattern := restToPattern(args)
	if pattern == "all" {
		p.SetAllUrlAction(act, dropResponseCode)
	} else if len(pattern) > 0 {
		p.SetUrlAction(pattern, act, dropResponseCode)
	}
}

func (d *DelayAction) EditCommand() string {
	switch d.Act {
	case DelayActNone:
		return "delay " + d.DurationCommand()
	case DelayActDelayEach:
		return "delay " + d.DurationCommand()
	case DelayActDropUntil:
		return "delay drop " + d.DurationCommand()
	default:
		return ""
	}
}

func (u *UrlProxyAction) EditCommand() string {
	switch u.Act {
	case UrlActNone:
		return "proxy"
	case UrlActCache:
		return "proxy cache"
	case UrlActDrop:
		return "proxy drop " + strconv.Itoa(u.DropResponseCode)
	default:
		return ""
	}
}

func (u *urlAction) EditCommand() string {
	c := ""
	if e := u.Act.EditCommand(); len(e) > 0 {
		c += e + " " + u.UrlPattern + "\n"
	}

	if e := u.Delay.EditCommand(); len(e) > 0 {
		c += e + " " + u.UrlPattern + "\n"
	}

	return c
}

func (u *urlAction) DeleteCommand() string {
	return "delete " + u.UrlPattern + "\n"
}

func commandDomainMode(p *Profile, mode, content string) {
	c, rest := cmd.TakeFirstArg(content)
	if c == "" {
		return
	}

	act := DomainActNone
	if mode != "redirect" && rest != "" {
		return
	}

	ip := ""
	if mode == "block" {
		act = DomainActBlock
	} else if mode == "redirect" {
		if rest != "" {
			addr := net.ParseIP(rest)
			if addr == nil {
				return
			} else {
				ip = addr.String()
			}
		}

		act = DomainActRedirect
	}

	if c == "all" {
		p.SetAllDomainAction(DomainAct(act), ip)
	} else {
		p.SetDomainAction(c, DomainAct(act), ip)
	}
}

func commandDomainDelete(p *Profile, content string) {
	c, rest := cmd.TakeFirstArg(content)
	if c == "" || rest != "" {
		return
	}

	if c == "all" {
		p.DeleteAllDomain()
	} else {
		p.DeleteDomain(c)
	}
}

func (d *DomainAction) EditCommand() string {
	switch d.Act {
	case DomainActNone:
		return "domain " + d.Domain + "\n"
	case DomainActBlock:
		return "domain block " + d.Domain + "\n"
	case DomainActRedirect:
		sep := ""
		if d.IP != "" {
			sep = " "
		}

		return "domain redirect " + d.Domain + sep + d.IP + "\n"
	default:
		return ""
	}
}

func (d *DomainAction) DeleteCommand() string {
	return "domain delete " + d.Domain + "\n"
}

func (p *Profile) ExportCommand() string {
	export := "# 此为客户端配置导出，可复制所有内容至“命令”输入窗口重新加载此配置 #\n\n"
	export += "# Name: " + p.Name + "\n"
	export += "# IP: " + p.Ip + "\n"
	export += "# Owner: " + p.Owner + "\n"

	export += "\n# 以下为 URL 命令定义 #\n"
	for _, u := range p.Urls {
		export += u.EditCommand()
	}

	export += "\n# 以下为域名命令定义 #\n"
	for _, d := range p.Domains {
		export += d.EditCommand()
	}

	export += "\n# end # \n"
	return export
}
