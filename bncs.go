package main

import (
	"fmt"
	"io"
	"log"
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

func (s *BncsSocket) SendSid_Auth_Info() {
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
	err := bncs.SendPacket(s.conn, 0x51)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func (s *BncsSocket) SendSid_LogonResponse2() {
	bncs := NewBncsPacket(nil)
	s.Bot.Bnls.SendBnls_HashData(s.Bot)
	for s.Bot.PasswordHash == nil {
		runtime.Gosched()
	}

	bncs.WriteDword(s.Bot.CdkeyData.ClientToken)
	bncs.WriteDword(s.Bot.CdkeyData.ServerToken)
	for i := 0; i < 5; i++ {
		bncs.WriteDword(s.Bot.PasswordHash[i])
	}
	bncs.WriteString(s.Bot.Config.Username)
	err := bncs.SendPacket(s.conn, 0x3a)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func (s *BncsSocket) SendSid_EnterChat() {
	bncs := NewBncsPacket(nil)
	bncs.WriteString(s.Bot.Config.Username)
	bncs.WriteString("")
	err := bncs.SendPacket(s.conn, 0x0a)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func (s *BncsSocket) SendSid_JoinChannel(channel string) {
	bncs := NewBncsPacket(nil)
	bncs.WriteDword(0x02)
	bncs.WriteString(channel)
	err := bncs.SendPacket(s.conn, 0x0c)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func (s *BncsSocket) handleSid_Auth_Info(bncs *BncsPacket) {
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
		log.Printf("[%s] Logon type: Broken Sha-1\n", s.Bot.ProfileName)
	case 0x01:
		log.Printf("[%s] Logon type: Nls version 1\n", s.Bot.ProfileName)
	case 0x02:
		log.Printf("[%s] Logon type: Nls version 2\n", s.Bot.ProfileName)
	default:
		log.Printf("[%s] Logon type: unknown [0x%x]\n", s.Bot.ProfileName, logonType)
	}

	s.Bot.Bnls.SendBnls_Cdkey(s.Bot, serverToken, s.Bot.Config.Cdkey)
	for s.Bot.CdkeyData == nil {
		runtime.Gosched()
	}
	s.Bot.CdkeyData.ServerToken = serverToken

	s.Bot.Bnls.SendBnls_VersionCheckEx2(s.Bot, s.Bot.CdkeyData.ClientToken, mpqFiletime, mpqFilename, valueString)
	for s.Bot.ExeInfo == nil {
		runtime.Gosched()
	}

	s.SendSid_Auth_Check()
}

func (s *BncsSocket) handleSid_Auth_Check(bncs *BncsPacket) {
	result := bncs.ReadDword()
	switch result {
	case 0x000:
		log.Printf("[%s] Passed authentication challenge\n", s.Bot.ProfileName)
		s.SendSid_LogonResponse2()
	case 0x100:
		log.Printf("[%s] Old game version\n", s.Bot.ProfileName)
	case 0x101:
		log.Printf("[%s] Invalid game version\n", s.Bot.ProfileName)
	case 0x102:
		log.Printf("[%s] Game version must be downgraded\n", s.Bot.ProfileName)
	case 0x200:
		log.Printf("[%s] Invalid cdkey\n", s.Bot.ProfileName)
	case 0x201:
		log.Printf("[%s] Cdkey in use\n", s.Bot.ProfileName)
	case 0x202:
		log.Printf("[%s] Banned cdkey\n", s.Bot.ProfileName)
	case 0x203:
		log.Printf("[%s] Cdkey for wrong product\n", s.Bot.ProfileName)
	}
}

func (s *BncsSocket) handleSid_LogonResponse2(bncs *BncsPacket) {
	result := bncs.ReadDword()
	switch result {
	case 0x00:
		log.Printf("[%s] Account login successful!\n", s.Bot.ProfileName)
		log.Printf("[%s] Entering home channel...\n", s.Bot.ProfileName)
		s.SendSid_EnterChat()
		s.SendSid_JoinChannel(s.Bot.Config.HomeChannel)
	case 0x01:
		log.Printf("[%s] Account does not exist\n", s.Bot.ProfileName)
	case 0x02:
		log.Printf("[%s] Invalid password\n", s.Bot.ProfileName)
	case 0x06:
		log.Printf("[%s] Account closed\n", s.Bot.ProfileName)
	default:
		log.Printf("Unknown logon response\n")
	}
}

func (s *BncsSocket) handleSid_ChatEvent(bncs *BncsPacket) {
	eventId := bncs.ReadDword()
	flags := bncs.ReadDword()
	ping := bncs.ReadDword()
	bncs.ReadDword() // Ip address (defunct)
	bncs.ReadDword() // Account number (defunct)
	bncs.ReadDword() // Registration authority (defunct)
	username := bncs.ReadString()
	text := bncs.ReadString()

	switch eventId {
	case 0x01: // Show user
		HandleEid_ShowUser(s.Bot, username, flags, ping, text)
	case 0x02: // User join
		HandleEid_UserJoin(s.Bot, username, flags, ping, text)
	case 0x03: // User leave
		HandleEid_UserLeave(s.Bot, username, flags, ping)
	case 0x04: // Received whisper
		HandleEid_RecvWhisper(s.Bot, username, flags, ping, text)
	case 0x05: // User talk
		HandleEid_UserTalk(s.Bot, username, flags, ping, text)
	case 0x09: // Flags update
		HandleEid_FlagUpdate(s.Bot, username, flags, ping, text)
	}
}

func (s *BncsSocket) recvLoop() {
	for {
		buf := make([]byte, 4096)
		n, err := s.conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("[%s] BncsSocket read problem : %s", s.Bot.ProfileName, err)
				break
			}
		}

		pos := n
		for pos > 4 {
			packetId := buf[1]
			packetLength := int(buf[2]) | int(buf[3]) << 8
			bncs := NewBncsPacket(buf[:packetLength])
			for len(buf) < packetLength {
				tmp := make([]byte, 4096)
				buf = append(buf, tmp...)
			}
			bncs.ReadByte()
			bncs.ReadByte()
			bncs.ReadWord()

			s.handlePacket(packetId, bncs)

			pos -= packetLength
			buf = buf[packetLength:]
		}

		runtime.Gosched()
	}
}

func (s *BncsSocket) handlePacket(id byte, bncs *BncsPacket) {
	switch id {
	case 0x0a:
	case 0x0f:
		s.handleSid_ChatEvent(bncs)
	case 0x25:
//		fmt.Printf("   >> Received SID_PING [%s]\n", s.Bot.ProfileName)
	case 0x3a:
//		fmt.Printf("   >> Received SID_LOGONRESPONSE2 [%s]\n", s.Bot.ProfileName)
		s.handleSid_LogonResponse2(bncs)
	case 0x50:
//		fmt.Printf("   >> Received SID_AUTHINFO [%s]\n", s.Bot.ProfileName)
		s.handleSid_Auth_Info(bncs)
	case 0x51:
//		fmt.Printf("   >> Received SID_AUTH_CHECK [%s]\n", s.Bot.ProfileName)
		s.handleSid_Auth_Check(bncs)
	default:
		log.Printf("[%s]  >> Received unknown packet!\n", s.Bot.ProfileName)
		bncs.Dump()
	}
}