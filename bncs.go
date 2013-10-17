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
	Server string
	Connected bool

	conn net.Conn
}

func NewBncsSocket() (bncsSocket *BncsSocket) {
	bncsSocket = new(BncsSocket)
	bncsSocket.Connected = false

	return bncsSocket
}

func (bncsSocket *BncsSocket) Connect(server string) (err error) {
	timeout := time.Duration(10) * time.Second
	bncsSocket.conn, err = net.DialTimeout("tcp", server, timeout)

	if err != nil {
		bncsSocket.Connected = false
		return err
	}
	
	bncsSocket.Server = bncsSocket.conn.RemoteAddr().String()
	bncsSocket.Connected = true

	return nil
}

func SendProtocolByte(bot *Bot) (err error) {
	_, err = bot.Bncs.conn.Write([]byte{1})

	return err
}

func SendSid_Auth_Info(bot *Bot) {
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

	err := bncs.SendPacket(bot.Bncs.conn, 0x50)
	if err != nil {
		fmt.Printf(err.Error())
	}

	go recvLoop(bot)
}

func SendSid_Auth_Check(bot *Bot) {
	bncs := NewBncsPacket(nil)
	bncs.WriteDword(bot.CdkeyData.ClientToken)
	bncs.WriteDword(bot.ExeInfo.Version)
	bncs.WriteDword(bot.ExeInfo.Checksum)
	bncs.WriteDword(0x01)
	bncs.WriteDword(0x00)
	bncs.WriteDword(bot.CdkeyData.KeyLength)
	bncs.WriteDword(bot.CdkeyData.ProductValue)
	bncs.WriteDword(bot.CdkeyData.PublicValue)
	bncs.WriteDword(0x00)
	for i := 0; i < 5; i++ {
		bncs.WriteDword(bot.CdkeyData.CdkeyHash[i])
	}
	bncs.WriteByteArray(bot.ExeInfo.StatString)
	bncs.WriteString("CookieBot4")

	err := bncs.SendPacket(bot.Bncs.conn, 0x51)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func SendSid_LogonResponse2(bot *Bot) {
	bncs := NewBncsPacket(nil)
	SendBnls_HashData(bot)
	for bot.PasswordHash == nil {
		runtime.Gosched()
	}
	bncs.WriteDword(bot.CdkeyData.ClientToken)
	bncs.WriteDword(bot.CdkeyData.ServerToken)
	for i := 0; i < 5; i++ {
		bncs.WriteDword(bot.PasswordHash[i])
	}
	bncs.WriteString(bot.Config.Username)
	
	err := bncs.SendPacket(bot.Bncs.conn, 0x3a)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func SendSid_EnterChat(bot *Bot) {
	bncs := NewBncsPacket(nil)
	bncs.WriteString(bot.Config.Username)
	bncs.WriteString("")
	
	err := bncs.SendPacket(bot.Bncs.conn, 0x0a)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func SendSid_JoinChannel(bot *Bot, channel string) {
	bncs := NewBncsPacket(nil)
	bncs.WriteDword(0x02)
	bncs.WriteString(channel)
	
	err := bncs.SendPacket(bot.Bncs.conn, 0x0c)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func handleSid_Auth_Info(bot *Bot, bncs *BncsPacket) {
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
		log.Printf("[%s] Logon type: Broken Sha-1\n", bot.ProfileName)
	case 0x01:
		log.Printf("[%s] Logon type: Nls version 1\n", bot.ProfileName)
	case 0x02:
		log.Printf("[%s] Logon type: Nls version 2\n", bot.ProfileName)
	default:
		log.Printf("[%s] Logon type: unknown [0x%x]\n", bot.ProfileName, logonType)
	}

	SendBnls_Cdkey(bot, serverToken, bot.Config.Cdkey)
	for bot.CdkeyData == nil {
		runtime.Gosched()
	}
	bot.CdkeyData.ServerToken = serverToken

	SendBnls_VersionCheckEx2(bot, mpqFiletime, mpqFilename, valueString)
	for bot.ExeInfo == nil {
		runtime.Gosched()
	}

	SendSid_Auth_Check(bot)
}

func handleSid_Auth_Check(bot *Bot, bncs *BncsPacket) {
	result := bncs.ReadDword()
	switch result {
	case 0x000:
		log.Printf("[%s] Passed authentication challenge\n", bot.ProfileName)
		SendSid_LogonResponse2(bot)
	case 0x100:
		log.Printf("[%s] Old game version\n", bot.ProfileName)
	case 0x101:
		log.Printf("[%s] Invalid game version\n", bot.ProfileName)
	case 0x102:
		log.Printf("[%s] Game version must be downgraded\n", bot.ProfileName)
	case 0x200:
		log.Printf("[%s] Invalid cdkey\n", bot.ProfileName)
	case 0x201:
		log.Printf("[%s] Cdkey in use\n", bot.ProfileName)
	case 0x202:
		log.Printf("[%s] Banned cdkey\n", bot.ProfileName)
	case 0x203:
		log.Printf("[%s] Cdkey for wrong product\n", bot.ProfileName)
	}
}

func handleSid_LogonResponse2(bot *Bot, bncs *BncsPacket) {
	result := bncs.ReadDword()
	switch result {
	case 0x00:
		log.Printf("[%s] Account login successful!\n", bot.ProfileName)
		log.Printf("[%s] Entering home channel...\n", bot.ProfileName)
		SendSid_EnterChat(bot)
		SendSid_JoinChannel(bot, bot.Config.HomeChannel)
	case 0x01:
		log.Printf("[%s] Account does not exist\n", bot.ProfileName)
	case 0x02:
		log.Printf("[%s] Invalid password\n", bot.ProfileName)
	case 0x06:
		log.Printf("[%s] Account closed\n", bot.ProfileName)
	default:
		log.Printf("Unknown logon response\n")
	}
}

func handleSid_ChatEvent(bot *Bot, bncs *BncsPacket) {
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
		HandleEid_ShowUser(bot, username, flags, ping, text)
	case 0x02: // User join
		HandleEid_UserJoin(bot, username, flags, ping, text)
	case 0x03: // User leave
		HandleEid_UserLeave(bot, username, flags, ping)
	case 0x04: // Received whisper
		HandleEid_RecvWhisper(bot, username, flags, ping, text)
	case 0x05: // User talk
		HandleEid_UserTalk(bot, username, flags, ping, text)
	case 0x07: // Channel information
		HandleEid_Channel(bot, text, flags)
	case 0x09: // Flags update
		HandleEid_FlagUpdate(bot, username, flags, ping, text)
	}
}

func recvLoop(bot *Bot) {
	for {
		buf := make([]byte, 4096)
		n, err := bot.Bncs.conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("[%s] BncsSocket read problem : %s", bot.ProfileName, err)
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

			handleBncsPacket(bot, packetId, bncs)

			pos -= packetLength
			buf = buf[packetLength:]
		}

		runtime.Gosched()
	}
}

func handleBncsPacket(bot *Bot, id byte, bncs *BncsPacket) {
	switch id {
	case 0x0a:
	case 0x0f:
		handleSid_ChatEvent(bot, bncs)
	case 0x25:
//		fmt.Printf("   >> Received SID_PING [%s]\n", s.Bot.ProfileName)
	case 0x3a:
//		fmt.Printf("   >> Received SID_LOGONRESPONSE2 [%s]\n", s.Bot.ProfileName)
		handleSid_LogonResponse2(bot, bncs)
	case 0x50:
//		fmt.Printf("   >> Received SID_AUTHINFO [%s]\n", s.Bot.ProfileName)
		handleSid_Auth_Info(bot, bncs)
	case 0x51:
//		fmt.Printf("   >> Received SID_AUTH_CHECK [%s]\n", s.Bot.ProfileName)
		handleSid_Auth_Check(bot, bncs)
	default:
		log.Printf("[%s]  >> Received unknown packet!\n", bot.ProfileName)
		bncs.Dump()
	}
}