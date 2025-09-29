package main

import (
	"log"

	"vigenere-cipher/internal"
)

func main() {
	cfg, err := internal.LoadConfig()
	if err != nil {
		log.Fatalln("Error loading Config:", err)
	}

	msg, addr := internal.ClientParseArgs(cfg)

	conn, err := addr.ClientSetup()
	if err != nil {
		log.Fatalln("Error connecting to Server:", err)
	}

	defer conn.Close()

	err = internal.Request(conn, cfg.BufferSize, *msg)
	if err != nil {
		log.Fatalln("Error sending request:", err)
	}
}
