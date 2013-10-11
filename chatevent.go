package main 

import (
	"fmt"
	"strings"
)

func HandleEid_ShowUser(bot *Bot, username string, flags int, ping int, text string) {
	fmt.Println(username)
}

func HandleEid_UserJoin(bot *Bot, username string, flags int, ping int, text string) {
	fmt.Println(username, "has joined the channel")
}

func HandleEid_UserLeave(bot *Bot, username string, flags int, ping int) {
	fmt.Println(username, "has left the channel")
}

func HandleEid_RecvWhisper(bot *Bot, username string, flags int, ping int, text string) {
	fmt.Println("From:", username, text)
}

func HandleEid_UserTalk(bot *Bot, username string, flags int, ping int, text string) {
	fmt.Println(username, ":", text)
	if strings.HasPrefix(text, bot.Config.Trigger) {
		CommandHandler(bot, text)
	}
}

func HandleEid_FlagUpdate(bot *Bot, username string, flags int, ping int, text string) {
}