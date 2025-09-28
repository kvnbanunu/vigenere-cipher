package main

import (
	"fmt"
	"log"
	"vigenere-cipher/internal/utils"
)

func main() {
	cfg, err := utils.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading Config: %v", err)
	}

	addr := utils.ServerParseArgs(cfg)

}
