package main 

import (
	"fmt"
	"strings"
)

func HandleEid_Channel(bot *Bot, channel string, flags int) {
	fmt.Println("join channel", channel)
	bot.UserList = NewUserList(channel)
}

func HandleEid_ShowUser(bot *Bot, username string, flags int, ping int, text string) {
	bot.UserList.AddUser(username, flags, ping, text[:4])
}

func HandleEid_UserJoin(bot *Bot, username string, flags int, ping int, text string) {
	bot.UserList.AddUser(username, flags, ping, text[:4])
}

func HandleEid_UserLeave(bot *Bot, username string, flags int, ping int) {
	bot.UserList.RemoveUser(username)
}

func HandleEid_RecvWhisper(bot *Bot, username string, flags int, ping int, text string) {
}

func HandleEid_UserTalk(bot *Bot, username string, flags int, ping int, text string) {
	if strings.HasPrefix(text, bot.Config.Trigger) {
		CommandHandler(bot, text)
	}
}

func HandleEid_FlagUpdate(bot *Bot, username string, flags int, ping int, text string) {
	user := bot.UserList.FindUsers(username)
	user[0].Flags = flags
}