package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("   ____            _    _      ____        _   ")
	fmt.Println("  / ___|___   ___ | | _(_) ___| __ )  ___ | |_ ")
	fmt.Println(" | |   / _ \\ / _ \\| |/ / |/ _ \\  _ \\ / _ \\| __|")
	fmt.Println(" | |__| (_) | (_) |   <| |  __/ |_) | (_) | |_ ")
	fmt.Println("  \\____\\___/ \\___/|_|\\_\\_|\\___|____/ \\___/ \\__|")
	fmt.Println("")

	log.Println("Initializing bot manager...")
	botManager := new(BotManager)
	if botManager == nil {
		log.Fatal("Failed to load bot manager!")
	}

	log.Println("Loading bots...")
	err := botManager.LoadBots(LoadConfig())
	if err != nil {
		log.Fatal(err.Error())
	} 

	bnls := NewBnlsSocket()
	server := "phix.no-ip.org:9367"
	log.Printf("[bnls] Connecting to %s...\n", server)
	err = bnls.Connect(server)
	if err != nil {
		fmt.Printf("[bnls] Failed to connect to %s\n", server)
		log.Fatal(err.Error())
	} else {
		log.Printf("[bnls] Successfully connected to %s\n", bnls.Server)
	}

	botManager.ConnectBots()

	for {
	}

} 