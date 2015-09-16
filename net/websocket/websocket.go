package websocket

import (
	"net"
)

func buildWebsocketHeader(path string, headers map[string][]string) []byte {
	s := ""
	s += "GET " + path + " HTTP/1.1\r\n"
	for h, vs := range headers {
		for _, v := range vs {
			s += h + ": " + v + "\r\n"
		}
	}

	s += "\r\n"
	return []byte(s)
}

func Conn(address, path string, headers map[string][]string) (*net.TCPConn, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return conn, err
	}

	bytes := buildWebsocketHeader(path, headers)
	for len(bytes) > 0 {
		n, err := conn.Write(bytes)
		bytes = bytes[n:]
		if err != nil {
			return conn, err
		}
	}

	return conn, nil
}
