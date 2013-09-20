package main

import (
	"net"
	"time"
)

type BnlsSocket struct {
	Server string
	Connected bool

	conn net.Conn
}

func NewBnlsSocket() (s *BnlsSocket) {
	s = new(BnlsSocket)
	s.Connected = false

	return s
}

func (s *BnlsSocket) Connect(server string) (err error) {
	timeout := time.Duration(30) * time.Second
	s.conn, err = net.DialTimeout("tcp", server, timeout)
	
	if err != nil {
		s.Connected = false
		return err
	}

	s.Server = s.conn.RemoteAddr().String()
	s.Connected = true

	return nil
}

