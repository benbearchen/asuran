package httpd

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Streaming struct {
	bind   *net.TCPAddr
	target *net.TCPAddr
	listen *net.TCPListener

	mutex sync.Mutex
}

func NewStreaming(bind, target string) (*Streaming, error) {
	s := new(Streaming)
	if len(target) > 0 {
		t, err := net.ResolveTCPAddr("tcp", target)
		if err != nil {
			return nil, err
		}

		s.target = t
	}

	b, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		return nil, err
	}

	listen, err := net.ListenTCP("tcp", b)
	if err != nil {
		return nil, err
	}

	s.bind = b
	s.listen = listen

	go s.run()

	return s, nil
}

func (s *Streaming) run() {
	for {
		conn, err := s.listen.AcceptTCP()
		if err != nil {
			log.Println("accept tcp fail:", err)
			continue
		}

		go func() {
			err := s.accept(conn)
			if err != nil {
				log.Println("deal conn fail:", err)
			}
		}()
	}
}

func (s *Streaming) SetTarget(target string) error {
	t, err := net.ResolveTCPAddr("tcp", target)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.target = t
	return nil
}

func (s *Streaming) getTarget() *net.TCPAddr {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.target
}

func (s *Streaming) accept(conn *net.TCPConn) error {
	defer conn.Close()

	target := s.getTarget()
	if target == nil {
		return fmt.Errorf("has no target")
	}

	t, err := net.DialTCP("tcp", nil, target)
	if err != nil {
		return err
	}

	defer t.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	var err2 error
	go func() {
		defer wg.Done()

		_, err2 = io.Copy(t, conn)
	}()

	_, err = io.Copy(conn, t)

	wg.Wait()

	if err == nil && err2 != nil {
		err = err2
	}

	return err
}
