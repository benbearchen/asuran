package profile

import (
	"strings"
)

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

func (p *Profile) ExportCommand() string {
	export := "# 此为客户端配置导出，可复制所有内容至“命令”输入窗口重新加载此配置 #\n\n"
	export += "# Name: " + p.Name + "\n"
	export += "# IP: " + p.Ip + "\n"
	export += "# Owner: " + p.Owner + "\n"

	if p.UrlDefault.Command() != "url " {
		export += "\n# URL 缺省配置\n" + p.UrlDefault.Command() + "\n"
	}

	export += "\n# 以下为 URL 命令定义 #\n"
	for _, u := range p.Urls {
		export += u.p.Command() + "\n"
	}

	export += p.ExportDNSCommand()

	export += "\n# end # \n"
	return export
}

func (p *Profile) ExportHistoryCommand(index int) (string, error) {
	return p.saver.LoadHistory(p.Ip, index)
}

func (p *Profile) ExportDNSCommand() string {
	export := "\n# 以下为域名命令定义 #\n"
	for _, d := range p.Domains {
		export += d.p.Command() + "\n"
	}
	return export
}
