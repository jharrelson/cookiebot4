package main

import (
	"fmt"
	"io"
//	"bytes"
//	"encoding/binary"
	"net"
	"runtime"
	"time"
)

type BncsSocket struct {
	Server string
	Connected bool

	conn net.Conn
}

func NewBncsSocket() (s *BncsSocket) {
	s = new(BncsSocket)
	s.Connected = false

	return s
}

func (s *BncsSocket) Connect(server string) (err error) {
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

func (s *BncsSocket) SendProtocolByte() (err error) {
	_, err = s.conn.Write([]byte{1})

	return err
}

func (s *BncsSocket) SendSid_AuthInfo() {
	bncs := NewBncsPacket(0x50)
	bncs.WriteDword(0x00)
	bncs.WriteDword(0x49583836)
	bncs.WriteDword(0x53544152)
	bncs.WriteDword(0xd9)
	bncs.WriteDword(0x00)
	bncs.WriteDword(0x00)
	bncs.WriteDword(0x00)
	bncs.WriteDword(0x00)
	bncs.WriteDword(0x00)
	bncs.WriteString("USA")
	bncs.WriteString("United States")
	err := bncs.SendPacket(s.conn)
	if err != nil {
		fmt.Printf(err.Error())
		bncs.Dump()
	}
	bncs.Dump()
	
	s.recvLoop()
}

func (s *BncsSocket) recvLoop() {
	for {
		buf := make([]byte, 4096)
		n, err := s.conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("BncsSocket read problem : %s", err)
			}
		}
		if n > 4 {
			buf = buf[:n]
			fmt.Printf("Read %d bytes\n", n)
			fmt.Printf("%x\n", buf)
		}

		runtime.Gosched()
	}
}