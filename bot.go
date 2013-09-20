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
}

type Bot struct {
	ProfileName string
	Config *BotConfig
	bncs *BncsSocket
}

func NewBot(profileName string, config *BotConfig) *Bot {
	b := new(Bot)
	b.ProfileName = profileName
	b.Config = config

	return b
}

func (b *Bot) Connect() bool {
	b.bncs = NewBncsSocket()
	log.Printf("[bncs] %s connecting to %s...", b.ProfileName, b.Config.Server)
	err := b.bncs.Connect(b.Config.Server)
	if err != nil {
		log.Printf("[bncs] %s failed to connect to %s [%s]", b.ProfileName, b.Config.Server, err.Error())
		return false
	}

	log.Printf("[bncs] %s successfully connected to %s", b.ProfileName, b.bncs.Server)
	log.Printf("[bncs] (%s) sending protocol byte (0x01)", b.ProfileName)
	err = b.bncs.SendProtocolByte()
	if err != nil {
		log.Printf("[bncs] (%s) failed to send protocol byte [%s]", b.ProfileName, err.Error())
		return false
	}
	b.bncs.SendSid_AuthInfo()
	
	return true
}

