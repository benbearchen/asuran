package net

import (
	"bufio"
	"fmt"
	gonet "net"
	"net/http"
)

func hijack(w http.ResponseWriter) (gonet.Conn, *bufio.ReadWriter, error) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("can't hijack %v", w)
	}

	return hj.Hijack()
}

func ResetResponse(w http.ResponseWriter) {
	conn, _, err := hijack(w)
	if err != nil {
		panic("panic for reset http.ResponseWriter")
	} else {
		conn.Close()
	}
}
