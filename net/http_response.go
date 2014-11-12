package net

import (
	"bufio"
	"fmt"
	"io"
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

type flushWriterWrapper struct {
	w *bufio.ReadWriter
}

func flushWriterWrap(w *bufio.ReadWriter) io.Writer {
	return &flushWriterWrapper{w}
}

func (w *flushWriterWrapper) Write(b []byte) (n int, err error) {
	return w.w.Write(b)
}

func (w *flushWriterWrapper) Flush() {
	w.w.Flush()
}

func TcpWriteHttp(w http.ResponseWriter, writeWrapper func(io.Writer) io.Writer, content []byte) bool {
	conn, writer, err := hijack(w)
	if err != nil {
		return false
	}

	defer conn.Close()
	defer writer.Flush()
	if writeWrapper != nil {
		writeWrapper(flushWriterWrap(writer)).Write(content)
	} else {
		writer.Write(content)
	}
	return true
}
