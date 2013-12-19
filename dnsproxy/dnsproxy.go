package dnsproxy

import (
	. "github.com/miekg/dns"

	"fmt"
	"net"
)

type DnsQuery interface {
	Query(clientIP, domain string) (string, []net.IP)
}

var (
	query DnsQuery
)

func root(w ResponseWriter, req *Msg) {
	if query == nil {
		w.Hijack()
		return
	}

	m := new(Msg)
	m.SetReply(req)

	domain := m.Question[0].Name
	host, _, err := net.SplitHostPort(w.RemoteAddr().String())
	if err != nil {
		host = ""
	}

	realDomain, ips := query.Query(host, domain)
	if ips == nil || len(ips) <= 0 {
		w.Hijack()
		return
	}

	m.Answer = make([]RR, len(ips))
	for i := 0; i < len(ips); i++ {
		m.Answer[i] = &A{Hdr: RR_Header{Name: realDomain, Rrtype: TypeA, Class: ClassINET, Ttl: 0}, A: ips[i]}
	}

	w.WriteMsg(m)
}

type defaultDnsQuery struct {
}

func (d *defaultDnsQuery) Query(clientIP, domain string) (string, []net.IP) {
	ips, err := querySystemDns(domain)
	if err != nil {
		return domain, nil
	}

	return domain, ips
}

func DnsProxy(q DnsQuery) {
	if q == nil {
		query = &defaultDnsQuery{}
	} else {
		query = q
	}

	HandleFunc(".", root)
	func() {
		err := ListenAndServe(":53", "udp", nil)
		if err != nil {
			fmt.Println("ListenAndServe: ", err.Error())
			panic(err)
		}
	}()
}
