package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"log"
//	"runtime"
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
		bot := NewBot(k, v)
		bm.bots.PushBack(bot)
	}

	if bm.bots.Len() < 1 {
		return fmt.Errorf("Problems loading config... could not load any bots")
	}

	log.Printf("Successfully loaded %d bots\n", bm.BotCount())

	return nil
}

func (bm *BotManager) ConnectBots(bnls *BnlsSocket) {
	log.Println("Connecting bots...")
	for b := bm.bots.Front(); b != nil; b = b.Next() {
		bot := b.Value.(*Bot)
		fmt.Println(" - Connecting", bot.ProfileName)
		fmt.Println(bot.Config)
		bot.Connect(bnls)
//		runtime.Gosched()
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
}

func (bm *BotManager) BotCount() int {
	return bm.bots.Len()
}