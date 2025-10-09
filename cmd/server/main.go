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

	server, err := socket.NewServer(addr)
	if err != nil {
		log.Fatal(err)
	}

	defer server.Cleanup()

	if err := server.Run(cfg.BufferSize, delay); err != nil {
		log.Fatal(err)
	}
}
