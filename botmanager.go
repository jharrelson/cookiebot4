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
	dbList *list.List
}

func (bm *BotManager) LoadBots(file []byte) (err error) {
	bm.bots = list.New()
	bm.dbList = list.New()

	var botMap map[string]BotConfig
	err = json.Unmarshal(file, &botMap)
	if err != nil {
		return err
	}

	for k, v := range botMap {
		bot := bm.NewBot(k, v)
		bm.bots.PushBack(bot)
	}

	if bm.bots.Len() < 1 {
		return fmt.Errorf("Problems loading config... could not load any bots")
	}

	log.Printf("Successfully loaded %d bots\n", bm.BotCount())

	return nil
}

func (bm *BotManager) NewBot(profileName string, config BotConfig) *Bot {
        b := new(Bot)
        b.ProfileName = profileName
        b.Config = config
	b.Database = LoadDatabase(bm.dbList, b.Config.Database)

        return b
}

func (bm *BotManager) ConnectBots(bnls *BnlsSocket) {
	log.Println("Connecting bots...")
	for b := bm.bots.Front(); b != nil; b = b.Next() {
		bot := b.Value.(*Bot)
		log.Println(" - Connecting", bot.ProfileName)
		bot.Connect(bnls)

		time.Sleep(time.Duration(100) * time.Millisecond)
	}
}

func (bm *BotManager) BotCount() int {
	return bm.bots.Len()
}