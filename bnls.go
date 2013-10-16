package main

import (
	"fmt"
	"io"
	"log"
	"net"
//	"runtime"
	"sync"
	"time"
)

type BnlsSocket struct {
	Server string
	Connected bool

	sync.Mutex
	conn net.Conn
}

func NewBnlsSocket() (s *BnlsSocket) {
	s = new(BnlsSocket)
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

	return nil
}

func SendBnls_Cdkey(bot *Bot, serverToken int, cdkey string) {
	bot.Bnls.Lock()

	bnls := NewBnlsPacket(nil)
	bnls.WriteDword(serverToken)
	bnls.WriteString(cdkey)

	err := bnls.SendPacket(bot.Bnls.conn, 0x01)
	if err != nil {
		fmt.Printf(err.Error())
	}

	recvBnlsPacket(bot)
}

func handleBnls_Cdkey(bot *Bot, bnls *BnlsPacket) {
	bot.CdkeyData = new(CdkeyData)

	bnls.ReadDword() // boolean result
	bot.CdkeyData.ClientToken = bnls.ReadDword()
	bot.CdkeyData.KeyLength = bnls.ReadDword()
	bot.CdkeyData.ProductValue = bnls.ReadDword()
	bot.CdkeyData.PublicValue = bnls.ReadDword()
	bnls.ReadDword() // Unknown 0
	bot.CdkeyData.CdkeyHash = make([]int, 5)
	for i := 0; i < 5; i++ {
		bot.CdkeyData.CdkeyHash[i] = bnls.ReadDword()
	}

	bot.Bnls.Unlock()
}

func SendBnls_VersionCheckEx2(bot *Bot, mpqFiletime []int, mpqFilename string, valueString []byte) {
	bot.Bnls.Lock()

	bnls := NewBnlsPacket(nil)
	bnls.WriteDword(0x01)
	bnls.WriteDword(0x00)
	bnls.WriteDword(0x00)
	bnls.WriteDword(mpqFiletime[0])
	bnls.WriteDword(mpqFiletime[1])
	bnls.WriteString(mpqFilename)
	bnls.WriteByteArray(valueString)

	err := bnls.SendPacket(bot.Bnls.conn, 0x1a)
	if err != nil {
		fmt.Printf(err.Error())
	}

	recvBnlsPacket(bot)
}

func handleBnls_VersionCheckEx2(bot *Bot, bnls *BnlsPacket) {
	bnls.ReadDword() // Success
	bot.ExeInfo = new(ExeInfo)
	bot.ExeInfo.Version = bnls.ReadDword()
	bot.ExeInfo.Checksum = bnls.ReadDword()
	bot.ExeInfo.StatString = bnls.ReadByteArray()
	bot.ExeInfo.Cookie = bnls.ReadDword()
	bot.ExeInfo.VersionByte = bnls.ReadDword()

	bot.Bnls.Unlock()
}

func SendBnls_HashData(bot *Bot) {
	bot.Bnls.Lock()

	bnls := NewBnlsPacket(nil)
	bnls.WriteDword(len(bot.Config.Password))
	bnls.WriteDword(0x02)
	bnls.WriteByteArray([]byte(bot.Config.Password))
	bnls.WriteDword(bot.CdkeyData.ClientToken)
	bnls.WriteDword(bot.CdkeyData.ServerToken)
	bnls.SendPacket(bot.Bnls.conn, 0x0b)

	recvBnlsPacket(bot)
}

func handleBnls_HashData(bot *Bot, bnls *BnlsPacket) {
	bot.PasswordHash = make([]int, 5)
	for i := 0; i < 5; i++ {
		bot.PasswordHash[i] = bnls.ReadDword()
	}
}

func recvBnlsPacket(bot *Bot) {
	buf := make([]byte, 4096)
	n, err := bot.Bnls.conn.Read(buf)
	if err != nil {
		if err != io.EOF {
			log.Printf("[bnls] BnlsSocket read problem : %s", err)
		}
	}
	if n > 4 {
		buf = buf[:n]
		bnls := NewBnlsPacket(buf)
		bnls.ReadWord() // Packet length
		packetId := bnls.ReadByte()
		handlePacket(bot, packetId, bnls)
	}
}
/*
func (s *BnlsSocket) recvLoop() {
	for {
		buf := make([]byte, 4096)
		n, err := s.conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("[bnls] BnlsSocket read problem : %s, err")
				break
			}
		}
		if n > 4 {
			buf = buf[:n]
			bnls := NewBnlsPacket(buf)
			bnls.ReadWord() // Packet length
			packetId := bnls.ReadByte()
			handlePacket(packetId, bnls)
		}

		runtime.Gosched()
	}
}
*/
func handlePacket(bot *Bot, id byte, bnls *BnlsPacket) {
	switch id {
	case 0x01:
		//fmt.Printf("BNLS: Received BNLS_CDKEY\n")
		handleBnls_Cdkey(bot, bnls)
	case 0x0b:
		//fmt.Printf("BNLS: Received BNLS_HASHDATA\n")
		handleBnls_HashData(bot, bnls)
	case 0x1a:
		//fmt.Printf("BNLS: Received BNLS_VERSIONCHECKEX2\n")
		handleBnls_VersionCheckEx2(bot, bnls)
	default:
		fmt.Printf("[bnls] Received unknown packet\n")
		bnls.Dump()
	}
}