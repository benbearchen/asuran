//go:build ignore

package main

import (
	"github.com/benbearchen/asuran/net/httpd"
	"github.com/benbearchen/asuran/net/udpd"

	"fmt"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("asuran_tunnel <target_ip>")
		return
	}

	ip := os.Args[1]

	go func() {
		port := "80"
		target := net.JoinHostPort(ip, port)

		s, err := httpd.NewStreaming(":"+port, target)
		if s == nil || err != nil {
			fmt.Println("failed to create http tunnel:", err)
			return
		}
	}()

	go func() {
		port := "53"
		target := net.JoinHostPort(ip, port)

		t, err := udpd.NewTunnel(":"+port, target)
		if t == nil || err != nil {
			fmt.Println("failed to create dns tunnel:", err)
		}
	}()

	for {
		str := ""
		fmt.Scanf("%s", &str)
		if str == "exit" {
			break
		}
	}
}
