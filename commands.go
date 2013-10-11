package main

import (
	"fmt"
	"strings"
)

func CommandHandler(bot *Bot, text string) {
	params := strings.Split(text, " ")
	command := params[0]
	command = command[len(bot.Config.Trigger):] // Trim the trigger off command
	command = strings.ToLower(command)

	switch command {
	case "ban":
		fmt.Println("ban", params[1:])
	case "unban":
		fmt.Println("unban", params[1])
	case "kick":
		fmt.Println("kick", params[1:])
	case "squelch":
		fallthrough
	case "ignore":
		fmt.Println("ignore")
	case "unsquelch":
		fallthrough
	case "unignore":
		fmt.Println("unignore")
	}
}