package main

import (
	"log"
//	"runtime"
)

type BotConfig struct {
	Username string
	Password string
	Cdkey string
	Server string
	HomeChannel string
	Trigger string
	Masters []string
	Database string
}

type Bot struct {
	ProfileName string
	Config BotConfig
	Database *Database

	CdkeyData *CdkeyData
	ExeInfo *ExeInfo
	PasswordHash []int

	Bncs *BncsSocket
	Bnls *BnlsSocket
}

type CdkeyData struct {
	ClientToken int
	ServerToken int
	KeyLength int
	ProductValue int
	PublicValue int
	CdkeyHash []int
}

type ExeInfo struct {
	Version int
	Checksum int
	StatString []byte
	Cookie int
	VersionByte int
}

func (bot *Bot) Connect(bnls *BnlsSocket) bool {
	bot.Bnls = bnls
	bot.Bncs = NewBncsSocket()

	log.Printf("[%s] Connecting to %s...", bot.ProfileName, bot.Config.Server)
	err := bot.Bncs.Connect(bot.Config.Server)
	if err != nil {
		log.Printf("[%s] Failed to connect to %s [%s]", 
			bot.ProfileName, bot.Config.Server, err.Error())
		return false
	}

	log.Printf("[%s] Successfully connected to %s", bot.ProfileName, bot.Bncs.Server)
	log.Printf("[%s] Sending protocol byte (0x01)", bot.ProfileName)
	err = SendProtocolByte(bot)
	if err != nil {
		log.Printf("[%s] Failed to send protocol byte [%s]", bot.ProfileName, err.Error())
		return false
	}
	SendSid_Auth_Info(bot)
	
	return true
}