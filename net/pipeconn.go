package net

import (
	_ "fmt"
	"io"
	gonet "net"
	"sync"
)

type writeCloser interface {
	CloseWrite() error
}

func PipeConn(a, b gonet.Conn) error {
	var wg sync.WaitGroup
	wg.Add(2)
	errchan := make(chan error)

	rw := func(w, r gonet.Conn) {
		defer wg.Done()
		defer func() {
			if c, ok := w.(writeCloser); ok {
				c.CloseWrite()
			} else {
				w.Close()
			}
		}()

		_, err := io.Copy(w, r)
		if err == io.EOF {
			err = nil
		}

		go func() {
			errchan <- err
		}()
	}

	go rw(a, b)
	go rw(b, a)

	wg.Wait()

	var err error = nil
	check := func(err2 error) {
		if err == nil && err2 != nil {
			err = err2
		}
	}

	check(<-errchan)
	check(<-errchan)
	close(errchan)

	check(a.Close())
	check(b.Close())

	return err
}
