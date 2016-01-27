package proxy

import (
	gonet "github.com/benbearchen/asuran/net"
	"github.com/benbearchen/asuran/policy"

	"bufio"
	_ "fmt"
	"io"
	"net"
	"net/http"
)

type chunkedWriter struct {
	chunked *policy.ChunkedPolicy
	w       io.Writer
	cache   []byte
	n       int
	sq      *policy.ChunkedSizeQueue
}

func newChunkedWriter(chunked *policy.ChunkedPolicy, w io.Writer) io.Writer {
	s := new(chunkedWriter)
	s.chunked = chunked
	s.w = w
	s.n = 0
	s.sq = nil

	switch chunked.Option() {
	case policy.ChunkedBlock:
		s.n = chunked.Block()
	case policy.ChunkedSize:
		s.sq = chunked.SizesQueue()
	}

	return s
}

func (t *chunkedWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return gonet.TryHijack(t.w)
}

func (t *chunkedWriter) Write(p []byte) (n int, err error) {
	t.Flush()
	if t.n > 0 {
		n := 0
		for i := 0; i < t.n && len(p) > 0; i++ {
			size := (len(p)+n)*(i+1)/t.n - n
			if size == 0 {
				size = 1
			}

			w, err := t.w.Write(p[:size])
			t.Flush()

			n += w
			if err != nil {
				return n, err
			}

			p = p[w:]
		}

		return n, nil
	} else if t.sq != nil {
		n := 0
		size := t.sq.Next()
		for len(p) > 0 {
			if size <= 0 || size > len(p) {
				size = len(p)
			}

			w, err := t.w.Write(p[:size])
			t.Flush()

			n += w
			if err != nil {
				return n, err
			}

			p = p[w:]
			size = t.sq.Next()
		}

		return n, nil
	} else {
		return t.w.Write(p)
	}
}

func (t *chunkedWriter) Flush() {
	if f, ok := t.w.(http.Flusher); ok {
		f.Flush()
	}
}
