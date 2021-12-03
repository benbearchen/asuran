package net

import "testing"

import (
	"io"
	"log"
	"net"
)

func TestPipeConn(t *testing.T) {
	b1 := ":22341"
	a1, err := net.ResolveTCPAddr("tcp", b1)
	if err != nil {
		t.Errorf("TestPipeConn() ResolveTCPAddr(%s) failed: %v", b1, err)
		return
	}

	b2 := ":22342"
	a2, err := net.ResolveTCPAddr("tcp", b2)
	if err != nil {
		t.Errorf("TestPipeConn() ResolveTCPAddr(%s) failed: %v", b2, err)
		return
	}

	s1, err := net.ListenTCP("tcp", a1)
	if err != nil {
		t.Errorf("TestPipeConn() Listen(%s) failed: %v", b1, err)
		return
	}

	defer s1.Close()
	go func() {
		for {
			c, err := s1.AcceptTCP()
			if err != nil {
				t.Errorf("TestPipeConn() AcceptTCP(%s) failed: %v", b1, err)
				return
			}

			defer c.Close()
			_, err = io.Copy(c, c)
			if err != nil && err != io.EOF {
				t.Errorf("TestPipeConn() AcceptTCP(%s).Copy(self) failed: %v", b1, err)
				return
			}

			log.Println("pipe server ok")

			break
		}
	}()

	s2, err := net.ListenTCP("tcp", a2)
	if err != nil {
		t.Errorf("TestPipeConn() Listen(%s) failed: %v", b2, err)
		return
	}

	defer s2.Close()
	go func() {
		for {
			c, err := s2.AcceptTCP()
			if err != nil {
				t.Errorf("TestPipeConn() Accept(%s) failed: %v", b2, err)
				return
			}

			c2, err := net.DialTCP("tcp", nil, a1)
			if err != nil {
				t.Errorf("TestPipeConn() DialTCP(%s) failed: %v", b1, err)
				return
			}

			err = PipeConn(c, c2)
			if err != nil {
				t.Errorf("TestPipeConn() PipeConn() failed: %v", err)
				return
			}

			log.Println("pipe ok")

			break
		}
	}()

	c1, err := net.DialTCP("tcp", nil, a2)
	if err != nil {
		t.Errorf("TestPipeConn() Connect(%s) failed: %v", b2, err)
		return
	}

	msg := "hello world"
	_, err = c1.Write([]byte(msg))
	if err != nil {
		t.Errorf("TestPipeConn() client.Write(%s) failed: %v", msg, err)
	}

	err = c1.CloseWrite()
	if err != nil {
		t.Errorf("TestPipeConn() client.CloseWrite() failed: %v", err)
	} else {
		log.Printf("client.CloseWrite()")
	}

	b, err := io.ReadAll(c1)
	if err != nil && err != io.EOF {
		t.Errorf("TestPipeConn() client.Read() failed: %v", err)
	}

	err = c1.CloseRead()
	if err != nil {
		t.Errorf("TestPipeConn() client.CloseRead() failed: %v", err)
	}

	if string(b) != msg {
		t.Errorf("TestPipeConn() return '%s' != '%s'", string(b), msg)
	}
}
