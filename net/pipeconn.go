package net

import (
	_ "fmt"
	"io"
	gonet "net"
	"sync"
)

func PipeConn(a, b gonet.Conn) error {
	var wg sync.WaitGroup
	wg.Add(2)
	errchan := make(chan error)

	rw := func(w, r gonet.Conn) {
		defer wg.Done()
		defer w.Close()
		_, err := io.Copy(w, r)
		if err != nil {
			go func() {
				errchan <- err
			}()
		}
	}

	go rw(a, b)
	go rw(b, a)

	wg.Wait()
	var err error = nil
	for e := range errchan {
		//fmt.Println(e)
		err = e
	}

	return err
}
