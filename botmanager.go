package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type BotManager struct {
	bots *list.List
}

func (bm *BotManager) LoadBots(file []byte) (err error) {
	bm.bots = list.New()

	var botMap map[string]BotConfig
	err = json.Unmarshal(file, &botMap)
	if err != nil {
		return err
	}
	
	for k, v := range botMap {
		bot := NewBot(k, &v)
		bm.bots.PushBack(bot)
	}

	if bm.bots.Len() < 1 {
		return fmt.Errorf("Problems loading config... could not load any bots")
	}

	log.Printf("Successfully loaded %d bots\n", bm.BotCount())

	return nil
}

func (bm *BotManager) ConnectBots() {
	log.Println("Connecting bots...")
	for b := bm.bots.Front(); b != nil; b = b.Next() {
		bot := b.Value.(*Bot)
		fmt.Println(" - Connecting", bot.ProfileName)
		go bot.Connect()
		time.Sleep(time.Duration(2) * time.Second)
	}
}

func (bm *BotManager) BotCount() int {
	return bm.bots.Len()
}