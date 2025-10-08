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

	msg, addr := utils.ClientParseArgs(cfg)

	var sock socket.SockAddr
	sock.Addr = addr

	conn, err := sock.ClientSetup()
	if err != nil {
		log.Fatalln("Error connecting to Server:", err)
	}

	defer conn.Close()

	err = socket.Request(conn, cfg.BufferSize, *msg)
	if err != nil {
		log.Fatalln("Error sending request:", err)
	}
}
