package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
)

type BnlsSocket struct {
	Server string
	Connected bool

	mutex *sync.Mutex
	bot *Bot
	conn net.Conn

	CdkeyDataChan chan *CdkeyData
}

func NewBnlsSocket() (s *BnlsSocket) {
	s = new(BnlsSocket)
	s.mutex = new(sync.Mutex)
	s.Connected = false

	return s
}

func (s *BnlsSocket) Connect(server string) (err error) {
	timeout := time.Duration(10) * time.Second
	s.conn, err = net.DialTimeout("tcp", server, timeout)
	
	if err != nil {
		s.Connected = false
		return err
	}

	s.Server = s.conn.RemoteAddr().String()
	s.Connected = true

	go s.recvLoop()

	return nil
}

func (s *BnlsSocket) SendBnls_Cdkey(bot *Bot, serverToken int, cdkey string) {
	s.mutex.Lock()
	s.bot = bot
	bnls := NewBnlsPacket(nil)
	bnls.WriteDword(serverToken)
	bnls.WriteString(cdkey)
	err := bnls.SendPacket(s.conn, 0x01)
	if err != nil {
		fmt.Printf(err.Error())
	}
//	bnls.Dump()
}

func (s *BnlsSocket) handleBnls_Cdkey(bnls *BnlsPacket) {
	s.bot.CdkeyData = new(CdkeyData)
	bnls.ReadDword() // boolean result
	s.bot.CdkeyData.ClientToken = bnls.ReadDword()
	s.bot.CdkeyData.KeyLength = bnls.ReadDword()
	s.bot.CdkeyData.ProductValue = bnls.ReadDword()
	s.bot.CdkeyData.PublicValue = bnls.ReadDword()
	bnls.ReadDword() // Unknown 0
	s.bot.CdkeyData.CdkeyHash = make([]int, 5)
	for i := 0; i < 5; i++ {
		s.bot.CdkeyData.CdkeyHash[i] = bnls.ReadDword()
	}
	s.mutex.Unlock()
}

func (s *BnlsSocket) SendBnls_VersionCheckEx2(bot *Bot, clientToken int, mpqFiletime []int, mpqFilename string, valueString []byte) {
	s.mutex.Lock()
	s.bot = bot
	bnls := NewBnlsPacket(nil)
	bnls.WriteDword(0x01)
	bnls.WriteDword(0x00)
	bnls.WriteDword(0x00)
	bnls.WriteDword(mpqFiletime[0])
	bnls.WriteDword(mpqFiletime[1])
	bnls.WriteString(mpqFilename)
	bnls.WriteByteArray(valueString)
	err := bnls.SendPacket(s.conn, 0x1a)
	if err != nil {
		fmt.Printf(err.Error())
	}
//	bnls.Dump()
}

func (s *BnlsSocket) handleBnls_VersionCheckEx2(bnls *BnlsPacket) {
	bnls.ReadDword() // Success
	s.bot.ExeInfo = new(ExeInfo)
	s.bot.ExeInfo.Version = bnls.ReadDword()
	s.bot.ExeInfo.Checksum = bnls.ReadDword()
	s.bot.ExeInfo.StatString = bnls.ReadByteArray()
	s.bot.ExeInfo.Cookie = bnls.ReadDword()
	s.bot.ExeInfo.VersionByte = bnls.ReadDword()
	s.mutex.Unlock()
}

func (s *BnlsSocket) recvLoop() {
	for {
		buf := make([]byte, 4096)
		n, err := s.conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("BnlsSocket read problem : %s, err")
				break
			}
		}
		if n > 4 {
			buf = buf[:n]
			fmt.Printf("BNLS: Read %d bytes\n", n)
//			fmt.Printf("%x\n", buf)

			bnls := NewBnlsPacket(buf)
//			bnls.Dump()
			bnls.ReadWord() // Packet length
			packetId := bnls.ReadByte()
			s.handlePacket(packetId, bnls)
		}

		runtime.Gosched()
	}
}

func (s *BnlsSocket) handlePacket(id byte, bnls *BnlsPacket) {
	switch id {
	case 0x01:
		fmt.Printf("BNLS: Received BNLS_CDKEY\n")
		s.handleBnls_Cdkey(bnls)
	case 0x1a:
		fmt.Printf("BNLS: Received BNLS_VERSIONCHECKEX2\n")
		s.handleBnls_VersionCheckEx2(bnls)
	default:
		fmt.Printf("BNLS: Received unknown packet\n")
	}
}