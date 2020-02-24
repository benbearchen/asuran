package udpd

import (
	"log"
	"net"
	"sync"
)

type Tunnel struct {
	conn   *net.UDPConn
	server *net.UDPAddr
	lock   sync.Mutex
	ss     map[string]*tunnelSession
}

type tunnelSession struct {
	conn   *net.UDPConn
	server *net.UDPAddr
	reply  func([]byte)
}

func NewTunnel(local, target string) (*Tunnel, error) {
	server, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		return nil, err
	}

	conn, err := bind(local)
	if err != nil {
		return nil, err
	}

	t := new(Tunnel)
	t.conn = conn
	t.server = server
	t.ss = make(map[string]*tunnelSession)

	go t.run()

	return t, nil
}

func (t *Tunnel) run() {
	defer t.conn.Close()

	for {
		b := make([]byte, 2048)
		n, remote, err := t.conn.ReadFromUDP(b)
		if err != nil {
			log.Println(err)
			continue
		}

		s, err := t.getSession(remote)
		if err != nil {
			log.Println(err)
			continue
		}

		if s != nil {
			s.post(b[:n])
		}
	}
}

func (t *Tunnel) getSession(remote *net.UDPAddr) (*tunnelSession, error) {
	addr := remote.String()

	t.lock.Lock()
	defer t.lock.Unlock()

	s, ok := t.ss[addr]
	if ok {
		return s, nil
	}

	conn, err := bind(":0")
	if err != nil {
		return nil, err
	}

	s = new(tunnelSession)
	s.conn = conn
	s.server = t.server
	s.reply = func(b []byte) {
		_, err := t.conn.WriteTo(b, remote)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("reply %x to %v\n", b, remote)
		}
	}

	go s.run()

	t.ss[addr] = s

	return s, nil
}

func (s *tunnelSession) run() {
	defer s.conn.Close()

	for {
		b := make([]byte, 2048)
		n, _, err := s.conn.ReadFromUDP(b)
		if err != nil {
			log.Println(err)
			continue
		}

		s.reply(b[:n])
	}
}

func (s *tunnelSession) post(b []byte) {
	_, err := s.conn.WriteTo(b, s.server)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("post %x to %v\n", b, s.server)
	}
}

func bind(bind string) (*net.UDPConn, error) {
	b, err := net.ResolveUDPAddr("udp", bind)
	if err != nil {
		return nil, err
	}

	return net.ListenUDP("udp", b)
}
