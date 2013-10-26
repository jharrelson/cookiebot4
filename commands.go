package main

import (
	"fmt"
	"strings"
//	"time"
)

type commandCallback func(bot *Bot, username string, params[] string, command string)

type commandStruct struct {
	reqFlags string
	callback commandCallback
}

var commandMap map[string]*commandStruct

func NewCommand(reqFlags string, callback commandCallback) *commandStruct {
	cmd := new(commandStruct)
	cmd.reqFlags = reqFlags
	cmd.callback = callback
	return cmd
}

func InitializeCommands() {
	commandMap = make(map[string]*commandStruct)
	commandMap["find"] = NewCommand("A", commandFind)
	commandMap["findex"] = NewCommand("A", commandFindex)
	commandMap["ban"] = NewCommand("O", commandBnet)
	commandMap["kick"] = NewCommand("O", commandBnet)
	commandMap["ignore"] = NewCommand("O", commandBnet)
	commandMap["squelch"] = NewCommand("O", commandBnet)
}

func CommandHandler(bot *Bot, username string, text string) {
	params := strings.Split(text, " ")
	command := params[0]
	params = params[1:]
	command = command[len(bot.Config.Trigger):] // Trim the trigger off command
	command = strings.ToLower(command)

	fmt.Println("params:", params)
	c := commandMap[command]
	if (c != nil) {
		if (bot.Database.EntryHasAny(username, c.reqFlags)) {
			c.callback(bot, username, params, command)
		} else {
			fmt.Println("not enough access")
		}
	} else {
		fmt.Println("command not found")
	}
}

func commandBnet(bot *Bot, username string, params[] string, command string) {
	if (len(params) < 1) {
		return
	} else if (len(params) < 2) {
		params = append(params, "")
	}

	var response string
	var extra string
	target := params[0]

	target = strings.ToLower(target)
	target = strings.Replace(target, ",", "", -1)
	target = strings.Replace(target, "@useast", "", -1)
	target = strings.Replace(target, "@uswest", "", -1)
	target = strings.Replace(target, "@asia", "", -1)
	target = strings.Replace(target, "@europe", "", -1)

	fmt.Println("target:", target)
	if (target == "") {
		SendSid_ChatCommand(bot, "Invalid arguments")
		return
	}

	for _, v := range params[1:] {
		extra += v
	}
	
	if (strings.Contains(target, "*")) {
		users := bot.UserList.FindUsers(target)
		for _, u := range users {
			if (extra == "") {
				entry := bot.Database.FindEntries(u.Name)
				if (entry != nil) {
					extra = entry[0].Comment
				}
			}

			if (!bot.Database.EntryHasAny(u.Name, "S")) {
				fmt.Printf("%s %s %s\n", command, u.Name, extra)
			}
		}
	} else {
		hasSafe := bot.Database.EntryHasAny(target, "S")
		if (hasSafe) {
			response = fmt.Sprintf("'%s' is safelisted", target)
		} else {
			if (extra == "") {
				entry := bot.Database.FindEntries(target)
				if (entry != nil) {
					extra = entry[0].Comment
				}
			}

			response = fmt.Sprintf("/%s %s %s", command, target, extra)
		}

		SendSid_ChatCommand(bot, response)
	}
}

func commandFindex(bot *Bot, username string, params[] string, command string) {
	if (len(params) < 1) {
		return
	} else if (len(params) < 2) {
		params = append(params, "")
	}

	entries := bot.Database.FindEntries(params[0])
	var response string

	// Mon Jan 2 15:04:05 MST 2006
	format := "Jan 2, 2006 @ 15:04:05"

	if (len(entries) == 0) {
		response = fmt.Sprintf("Unable to find any database entries related to '%s'", params[0])
	} else {
		entry := entries[0]
		var access string

		if (strings.HasPrefix(params[0], "%")) {
			access = IntToFlags(entry.Access.(uint32))
		} else {
			switch (entry.Access.(type)) {
			case string:
				groups := bot.Database.FindEntries(entry.Access.(string))
				if (groups == nil) {
					access = fmt.Sprintf("( %s : <none> )")
				} else {
					access = fmt.Sprintf("( %s : %s )", entry.Access.(string), IntToFlags(groups[0].Access.(uint32)))
				}
			case uint32:
				access = IntToFlags(entry.Access.(uint32))
			}
		}
		response = fmt.Sprintf("%s => %s ", entry.Name, access)

		switch (params[1]) {
		case "-added", "-created":
			response += fmt.Sprintf("[Added: %s by %s]", entry.CreatedDate.Format(format), entry.CreatedBy) 
		case "-edited", "-modified":
			if (entry.ModifiedDate.Unix() > 0) {
				response += fmt.Sprintf("[Modified: %s by %s]", entry.ModifiedDate.Format(format), entry.ModifiedBy)
			} else {
				response += fmt.Sprintf("[Modified: <never> by <never>")
			}
		case "-comment", "":
			if (entry.Comment != "") {
				response += fmt.Sprintf("[Comment: %s]", entry.Comment)
			}
		}
	}

	SendSid_ChatCommand(bot, response)
}

func commandFind(bot *Bot, username string, params[] string, command string) {
	if (len(params) < 1) {
		return
	}

	entries := bot.Database.FindEntries(params[0])
	var response string

	if (len(entries) == 0) {
		response = fmt.Sprintf("Unable to find any database entries related to '%s'", params[0])
	} else {
		if (strings.HasPrefix(params[0], "%")) {
			response = fmt.Sprintf("%d Group(s): ", len(entries))
			for _, e := range entries {
				if (e.Access.(uint32) == 0) {
					response += fmt.Sprintf("%s ( <none> ), ", e.Name)
				} else {
					response += fmt.Sprintf("%s ( %s ), ", e.Name, IntToFlags(e.Access.(uint32)))
				}
			}
		} else {
			response = fmt.Sprintf("%d User(s): ", len(entries))
			for _, e := range entries {
				switch (e.Access.(type)) {
				case string:
					groups := bot.Database.FindEntries(e.Access.(string))
					if (groups == nil) {
						response += fmt.Sprintf("%s ( %s : <none> ), ", e.Name, e.Access.(string))
					} else {
						response += fmt.Sprintf("%s ( %s : %s ), ", e.Name, e.Access.(string), IntToFlags(groups[0].Access.(uint32)))
					}
				case uint32:
					response += fmt.Sprintf("%s ( %s ), ", e.Name, IntToFlags(e.Access.(uint32)))
				}
			}
		}

		//response = strings.TrimSuffix(response, ", ")
	}

	SendSid_ChatCommand(bot, response)
}