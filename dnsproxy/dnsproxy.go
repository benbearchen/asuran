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
	if ips == nil {
		w.Hijack()
		return
	}

	if len(ips) > 0 {
		m.Answer = make([]RR, len(ips))
		for i := 0; i < len(ips); i++ {
			if ip4 := ips[i].To4(); ip4 != nil {
				m.Answer[i] = &A{Hdr: RR_Header{Name: realDomain, Rrtype: TypeA, Class: ClassINET, Ttl: 0}, A: ip4}
			} else {
				m.Answer[i] = &AAAA{Hdr: RR_Header{Name: realDomain, Rrtype: TypeAAAA, Class: ClassINET, Ttl: 0}, AAAA: ips[i]}
			}
		}
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
