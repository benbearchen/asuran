package profile

import (
	"strings"
)

func CommandUsage() string {
	return `commands:
-------
# 以 # 开头的行为注释

url [(set|update)] [settings...] (<url-pattern>|all)

settings... ::=
      [drop <duration>]
      [(delay|timeout) [body] [rand] <duration>]
      [(proxy|cache|status <responseCode>|(map|redirect) <resource-url>|rewrite <url-encoded-content>|restore <store-id>|tcpwrite <url-encoded-content>)]
      [speed <speeds>]
      [(dont302|do302)]
      [(disable304|allow304)]
      [content-type (default|remove|empty|<content-type>)]
      [host <ip:port>]


url delete (<url-pattern>|all)

domain ([default]|block|proxy|null) (delay [rand] <duration>) (<domain-name>|all) [<ip>]

domain delete (<domain-name>|all)


compatible commands:
-------
<ip> <domain-name>
# 等价于  domain <domain-name> <ip>


-------
注：
* <> 尖括号表示参数一定要替换成实际值，否则出错
* [] 中括号表示参数可有可无
* (a|b|...) 表示 a 或 b 等等多选一
* 下面注释以“**”开始的行，表示未实现功能
-------

url command:
        url 命令表示按 url-pattern 匹配、操作 HTTP 请求。
        下面为参数说明：


              下面设置模式只能二选一：
              [默认] update
    set       命令中出现的时间或内容模式会设置，未出现的模式设置成默认值。
    update    仅设置命令中出现的时间或内容模式；未出现的模式不变。


              下面时间模式只能多选一：
	      [默认] delay 0
    delay <duration>
              所有请求延时 duration 才开始返回；
              duration == 0 表示不延时，立即返回。
    drop <duration>
              从 URL 第一次至 duration 时间内的请求一律丢弃，
              直到 duration 以后的请求正常返回。
              duration == 0 表示只丢弃第一次请求。
              被 drop 将无法响应 cache、status 等其它请求。
              ** “丢弃”的效果可能无法很好实现 **
    timeout <duration>
              所有请求等待 duration 时间后，丢弃请求。

              时间可选参数：
    body      对 HTTP 回复的 body（而不是 headers）进行延时或超时处理。
              允许对 headers 和 body 独立设置。
              特殊地，timeout body 在发送进行 duration 时长后断开链接。
    rand      不使用固定时长，而是随机生成 [0, 1) * duration。


              下面几种内容模式只能多选一：
	      [默认] proxy
    proxy     代理 URL 请求结果。
    cache     缓存源 URL 请求结果，下次请求起从缓存返回。
    status <responseCode>
              对请求直接以 responseCode 回应。
              responseCode 可以是 404、502 等，
              但不应该是 200、302 等。
    map <resource-url>
              代理将请求 resource-url 的内容并返回。
    redirect <resource-url>
              返回 302 以让客户端自己跳转至 resource-url。
    rewrite <url-encoded-content>
              以 url-encoded-content 的原始内容返回。
    restore <store-id>
              以预先保存的名字为 store-id 的内容返回。
              store-id 内容可以上传，也可以从请求历史修改。
    tcpwrite <url-encoded-content>
              直接以 TCP 而不是 HTTP 格式返回内容


    speed <speeds>
              限制回复带宽最高为 speeds，默认单位为 B/s，
              即字节每秒。支持 GB, MB, KB 等量纲。
              如 100, 99KB, 0.5MB/s 均可。


    dont302, do302
              决定是否由 asuran 来执行 302 跳转，二选一。[默认] dont302
              dont302 可以让客户端收到 302 跳转；
              另外，目标服务器返回的 301、307 会被改写为 302。
              do302 则会让 asuran 去直接访问 302 后的链接，
              且 asuran 支持多次 302 跳转。

    disable304, allow304
              决定是否允许服务器返回 304，二选一。[默认] allow304
              不允许则请求时删除 If-None-Match 与 If-Modified-Since。


    content-type default
    content-type remove
    content-type empty
    content-type <content-type> 
              处理回复的 Content-Type。[默认] default
              default 表示不做任何处理；
              remove 表示移除回复的 Content-Type；
              empty 表示将回复的 Content-Type 置为空；
              其它将 Content-Type 设置为 <content-type> 值。
              <content-type> 不能包含空格，所以可能不支持 multipart。


    host <ip:port>
              指定实际连接的服务器地址


    delete    删除对 url-pattern 的配置。

    <duration>
              时长，可选单位：ms, s, m, h。默认为 s
              例：90s 或 1.5m
    <responseCode>
              HTTP 返回状态码，如 200/206、302/304、404 等。
    <resource-url>
              外部资源的 URL 地址（http:// 啥的）。
    <url-encoded-content>
              以 url-encoded 方式编码的文本或者二进制内容。
              直接返回给客户端。
    <store-id>
              上传内容或者修改请求历史内容，得到内容的 id。
              id 对应内容可方便修改。
    <url-pattern>
              [domain[:port]]/[path][?key=value]
              分域名[端口]、根路径与查询参数三种匹配。
              域名忽略则匹配所有域名。
              根路径可以匹配到目录或文件。
              查询参数匹配时忽略顺序，但列出参数必须全有。
              域名支持通配符“*”，如 *.com, *play.org
    all
              特殊地，all 可以操作所有已经配置的 url-pattern。

domain mode:
              以下域名模式只能多选一：
    [default] 域名默认为正常通行，返回自定义或正常结果。
              返回自定义 <ip> 如果有设置；否则实时查询后返回。
    block     屏蔽域名，不返回任何结果。
    proxy     返回 asuran IP，以代理设备 HTTP 请求。
    null      返回查询无结果

    delay [rand] <duration>
              延时后返回，定义与 url delay 相同（不支持 body）

<domain-name>:
    ([^.]+.)+[^.]+
              域名，目前支持英文域名（中文域名未验证）。
    all
              特殊地，all 可以操作所有已经配置的域名。

<ip>:         自定义域名的 IP 地址，比如 192.168.1.3。
              不设置则在需要返回 IP 时由 asuran 查询实际 IP。


-------
examples:

url delay 5s g.cn/search

url github.com/?cmd=1

url cache golang.org/doc/code.html

url status 404 baidu.com/

domain g.cn

domain block baidu.com

domain proxy g.cn

domain delete g.cn
`
}

func UrlToPattern(url string) string {
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

func (u *urlAction) EditCommand() string {
	command := u.p.Command()
	if command != "" {
		command += "\n"
	}

	return command
}

func (u *urlAction) DeleteCommand() string {
	return "url delete " + u.UrlPattern + "\n"
}

func (d *DomainAction) EditCommand() string {
	command := d.p.Command()
	if command != "" {
		command += "\n"
	}

	return command
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

	export += p.ExportDNSCommand()

	export += "\n# end # \n"
	return export
}

func (p *Profile) ExportDNSCommand() string {
	export := "\n# 以下为域名命令定义 #\n"
	for _, d := range p.Domains {
		export += d.EditCommand()
	}
	return export
}
