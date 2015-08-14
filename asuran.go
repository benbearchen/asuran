package main

import (
	"github.com/benbearchen/asuran/dnsproxy"
	"github.com/benbearchen/asuran/net"
	"github.com/benbearchen/asuran/profile"
	"github.com/benbearchen/asuran/util/cmd"
	"github.com/benbearchen/asuran/web/proxy"

	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func usage() {
	fmt.Println(`web transparent proxy

proxy test:
  http://localhost/test/target.domain:port/path
        return content of 'http://target.domain:port/path' & runtime info
  http://localhost/to/target.domain:port/path
        proxy 'http://target.domain:port/path', fetch & return content

cmd:
  bench [target.domain:port/path]
        benchmark url http://target.domain:port/path in some goroutines,
        finally calc used/avg time.  default: http://localhost

  delete profile <ip>
`)
}

func bench(target string) {
	fmt.Println("begin bench to: " + target + " ...")

	succ := 0
	fail := 0
	var t time.Duration = 0
	for i := 0; i < 10 || t.Seconds() < 30; i++ {
		start := time.Now()
		resp, err := net.NewHttpGet(target)
		if err == nil {
			defer resp.Close()
			_, err := resp.ReadAll()
			if err == nil {
				succ++
				end := time.Now()
				t += end.Sub(start)
				continue
			}
		}

		fail++
	}

	fmt.Printf("succ: %d\n", succ)
	if succ > 0 {
		fmt.Println("used: " + t.String())
		fmt.Println("avg:  " + time.Duration(int64(t)/int64(succ)).String())
	}

	fmt.Printf("fail: %d\n", fail)
}

func benchN(target string) {
	c := make(chan int)
	n := 4
	for i := 0; i < n; i++ {
		go func() { bench(target); c <- 1 }()
	}

	for i := 0; i < n; i++ {
		<-c
	}
}

const VersionCode = "0.2.4-dev"

func version() {
	fmt.Println(`asuran ` + VersionCode + `, a web proxy with dns
`)
}

var (
	nodns = flag.Bool("nodns", false, "nodns DISABLE the dns function")
)

func main() {
	version()

	flag.Parse()

	p := proxy.NewProxy(VersionCode)

	ipProfiles := profile.NewIpProfiles()
	ipProfiles.BindProxyHostOperator(p.NewProxyHostOperator())
	ipProfiles.SetDefaultCopyProfile("localhost")

	p.BindUrlOperator(ipProfiles.OperatorUrl())
	p.BindProfileOperator(ipProfiles.OperatorProfile())
	p.BindDomainOperator(ipProfiles.OperatorDomain())

	if *nodns {
		p.DisableDNS()
	} else {
		go dnsproxy.DnsProxy(dnsproxy.NewPolicy(p.NewDomainOperator()))
	}

	var c cmd.Command
	c.OpenConsole()
	for {
		fmt.Print("\n$ ")
		command := c.Read()
		command, rest := cmd.TakeFirstArg(command)
		switch command {
		case "":
		case "exit":
			return
		case "usage":
			fallthrough
		case "help":
			usage()
		case "version":
			version()
		case "bench":
			if len(rest) == 0 {
				fmt.Println("usage: bench <url>")
			} else {
				url := "http://"
				if strings.HasPrefix(rest, url) {
					url = rest
				} else {
					url += rest
				}

				benchN(url)
			}
		case "bind":
			port, err := strconv.Atoi(rest)
			if err != nil {
				fmt.Println("usage: bind <port>\nport: in 1~65535")
			} else {
				fmt.Println("")
				bindNew := p.Bind(port)
				if bindNew {
					fmt.Println("port", port, "binds ok")
				} else {
					fmt.Println("port had already bound")
				}
			}
		case "delete":
			mod, rest := cmd.TakeFirstArg(rest)
			switch mod {
			case "profile":
				if ipProfiles.Delete(rest) {
					fmt.Println(`profile "` + rest + `" deleted`)
				} else {
					fmt.Println(`profile "` + rest + `" don't exist`)
				}
			default:
				usage()
			}
		default:
			usage()
			fmt.Println(`UNKNOWN command: "` + command + `" ` + rest)
		}
	}
}
