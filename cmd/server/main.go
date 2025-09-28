package main

import (
	"log"
	"time"

	"vigenere-cipher/internal"
)

func main() {
	cfg, err := internal.LoadConfig()
	if err != nil {
		log.Fatalln("Error loading Config:", err)
	}

	addr := internal.ServerParseArgs(cfg)
	
	fd, err := addr.ServerSetup()
	if err != nil {
		log.Fatalln("Error setting up Server:", err)
	}

	defer fd.Close()

	f := internal.Flag{Exit: false}

	internal.HandleSignal(fd, &f)

	for {
		conn, err := fd.AcceptTCP()
		if err != nil {
			if f.Exit {
				time.Sleep(time.Millisecond)
				continue
			}

			log.Println("Error accepting connection:", err)
			break
		}
		err = internal.HandleConnection(conn, cfg.BufferSize)
		if err != nil {
			log.Println("Error handling connection:", err)
			break
		}
	}
}
