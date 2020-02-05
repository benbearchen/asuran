// +build http proxy

package main

import (
	"github.com/benbearchen/asuran/net/httpd"

	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("tunnel :<localport> <target:port>")
		return
	}

	s, err := httpd.NewStreaming(os.Args[1], os.Args[2])
	if s == nil || err != nil {
		fmt.Println("failed to create tunnel:", err)
		return
	}

	for {
		str := ""
		fmt.Scanf("%s", &str)
		if str == "exit" {
			break
		}
	}
}
