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

	payload, addr := utils.ClientParseArgs(cfg)

	conn, err := socket.ClientSetup(addr)
	if err != nil {
		log.Fatalln("Error connecting to Server:", err)
	}

	defer conn.Close()

	err = socket.SendRequest(conn, cfg.BufferSize, *payload)
	if err != nil {
		log.Fatalln("Error sending request:", err)
	}
}
