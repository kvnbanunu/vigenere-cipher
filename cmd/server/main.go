package main

import (
	"log"

	"vigenere-cipher/internal/socket"
	"vigenere-cipher/internal/utils"
)

func main() {
	cfg, err := utils.LoadConfig()
	if err != nil {
		log.Fatalln("Error loading Config:", err)
	}

	addr, delay := utils.ServerParseArgs(cfg)

	server, err := socket.ServerSetup(addr)
	if err != nil {
		log.Fatalln("Error setting up Server:", err)
	}

	// cleanup will always happen when main returns
	defer server.Cleanup()

	socket.HandleSignal(server)

	if err := server.Run(cfg.BufferSize, delay); err != nil {
		log.Println(err)
	}
}
