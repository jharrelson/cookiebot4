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
	Bot *Bot
	Server string
	Connected bool

	conn net.Conn
}

func NewBncsSocket(b *Bot) (s *BncsSocket) {
	s = new(BncsSocket)
	s.Bot = b
	s.Connected = false

	return s
}

func (s *BncsSocket) Connect(server string) (err error) {
	timeout := time.Duration(10) * time.Second
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
	bncs := NewBncsPacket(nil)
	bncs.WriteDword(0x00)
	bncs.WriteDword(0x49583836)
	bncs.WriteDword(0x53544152)
	bncs.WriteDword(0xd3)
	bncs.WriteDword(0x00)
	bncs.WriteDword(0x00)
	bncs.WriteDword(0x00)
	bncs.WriteDword(0x00)
	bncs.WriteDword(0x00)
	bncs.WriteString("USA")
	bncs.WriteString("United States")
	err := bncs.SendPacket(s.conn, 0x50)
	if err != nil {
		fmt.Printf(err.Error())
	}
	//bncs.Dump()

	go s.recvLoop()
}

func (s *BncsSocket) SendSid_Auth_Check() {
	bncs := NewBncsPacket(nil)
	bncs.WriteDword(s.Bot.CdkeyData.ClientToken)
	bncs.WriteDword(s.Bot.ExeInfo.Version)
	bncs.WriteDword(s.Bot.ExeInfo.Checksum)
	bncs.WriteDword(0x01)
	bncs.WriteDword(0x00)
	bncs.WriteDword(s.Bot.CdkeyData.KeyLength)
	bncs.WriteDword(s.Bot.CdkeyData.ProductValue)
	bncs.WriteDword(s.Bot.CdkeyData.PublicValue)
	bncs.WriteDword(0x00)
	for i := 0; i < 5; i++ {
		bncs.WriteDword(s.Bot.CdkeyData.CdkeyHash[i])
	}
	bncs.WriteByteArray(s.Bot.ExeInfo.StatString)
	bncs.WriteString("CookieBot4")
	fmt.Printf("  >> Sending SID_AUTH_CHECK [%s]\n", s.Bot.ProfileName)
	err := bncs.SendPacket(s.conn, 0x51)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func (s *BncsSocket) handleSid_AuthInfo(bncs *BncsPacket) {
	logonType := bncs.ReadDword()
	serverToken := bncs.ReadDword()
	bncs.ReadDword() // Udp value
	mpqFiletime := make([]int, 2)
	mpqFiletime[0] = bncs.ReadDword()
	mpqFiletime[1] = bncs.ReadDword()
	mpqFilename := bncs.ReadString()
	valueString := bncs.ReadByteArray()

	switch logonType {
	case 0x00:
		fmt.Printf("Logon type: Broken Sha-1\n")
	case 0x01:
		fmt.Printf("Logon type: Nls version 1\n")
	case 0x02:
		fmt.Printf("Logon type: Nls version 2\n")
	default:
		fmt.Printf("Logon type: unknown [0x%x]\n", logonType)
	}

	s.Bot.Bnls.SendBnls_Cdkey(s.Bot, serverToken, s.Bot.Config.Cdkey)
	for s.Bot.CdkeyData == nil {
		runtime.Gosched()
	}
//	fmt.Println("Got cdkey data")
	s.Bot.Bnls.SendBnls_VersionCheckEx2(s.Bot, s.Bot.CdkeyData.ClientToken, mpqFiletime, mpqFilename, valueString)
	for s.Bot.ExeInfo == nil {
		runtime.Gosched()
	}
//	fmt.Println("Got exe info")

	s.SendSid_Auth_Check()
}

func (s *BncsSocket) handleSid_Auth_Check(bncs *BncsPacket) {
	result := bncs.ReadDword()
	fmt.Printf("\n@@@@@@@@@@@@@@@@@@@@@ %s\n", s.Bot.ProfileName)
	switch result {
	case 0x000:
		fmt.Printf("Passed challenge\n")
	case 0x100:
		fmt.Printf("Old game version\n")
	case 0x101:
		fmt.Printf("Invalid game version\n")
	case 0x102:
		fmt.Printf("Game version must be downgraded\n")
	case 0x200:
		fmt.Printf("Invalid cdkey\n")
	case 0x201:
		fmt.Printf("Cdkey in use\n")
	case 0x202:
		fmt.Printf("Banned cdkey\n")
	case 0x203:
		fmt.Printf("Cdkey for wrong product\n")
	}
}

func (s *BncsSocket) recvLoop() {
	for {
		fmt.Println(s.Bot.ProfileName)
		buf := make([]byte, 4096)
		n, err := s.conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("BncsSocket read problem : %s", err)
				break
			}
		}
		if n > 4 {
			buf = buf[:n]
			fmt.Printf("Read %d bytes\n", n)
//			fmt.Printf("%x\n", buf)

			bncs := NewBncsPacket(buf)
//			bncs.Dump()
			bncs.ReadByte() // FF
			packetId := bncs.ReadByte()
			bncs.ReadWord() // Packet Length
//			packetLength := bncs.ReadWord()
			s.handlePacket(packetId, bncs)
		}

		runtime.Gosched()
	}
}

func (s *BncsSocket) handlePacket(id byte, bncs *BncsPacket) {
	switch id {
	case 0x25:
		fmt.Printf("   >> Received SID_PING [%s]\n", s.Bot.ProfileName)
	case 0x50:
		fmt.Printf("   >> Received SID_AUTHINFO [%s]\n", s.Bot.ProfileName)
		s.handleSid_AuthInfo(bncs)
	case 0x51:
		fmt.Printf("   >> Received SID_AUTH_CHECK [%s]\n", s.Bot.ProfileName)
		s.handleSid_Auth_Check(bncs)
	default:
		fmt.Printf("   >> Received unknown packet!")
	}
}