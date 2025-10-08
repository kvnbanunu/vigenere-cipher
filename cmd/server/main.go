package main

import (
	"log"
	"time"

	"vigenere-cipher/internal/socket"
	"vigenere-cipher/internal/utils"
)

func main() {
	cfg, err := utils.LoadConfig()
	if err != nil {
		log.Fatalln("Error loading Config:", err)
	}

	var sock socket.SockAddr
	sock.Addr = utils.ServerParseArgs(cfg)

	fd, err := sock.ServerSetup()
	if err != nil {
		log.Fatalln("Error setting up Server:", err)
	}

	defer fd.Close()

	f := socket.Flag{Exit: false}

	socket.HandleSignal(fd, &f)

	for !f.Exit {
		conn, err := fd.AcceptTCP()
		if err != nil {
			if f.Exit {
				time.Sleep(time.Millisecond)
				continue
			}

			log.Println("Error accepting connection:", err)
			break
		}
		go socket.HandleConnection(conn, cfg.BufferSize)
	}
}
