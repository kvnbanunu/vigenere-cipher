package socket

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"vigenere-cipher/internal/utils"
	"vigenere-cipher/internal/vigenere"
)

type Flag struct {
	Exit bool
}

func ServerSetup(addr *utils.Addr) (*net.TCPListener, int, error) {
	addrStr := fmt.Sprintf("%s:%s", addr.IP, addr.Port)

	sock, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		return nil, 0, err
	}

	listener, err := net.ListenTCP("tcp", sock)
	if err != nil {
		return nil, 0, err
	}

	fmt.Println("Now listening on ", addrStr)

	fd, err := listener.File()
	if err != nil {
		listener.Close()
		return nil, 0, err
	}

	return listener, int(fd.Fd()), nil
}

func ClientSetup(addr *utils.Addr) (*net.TCPConn, error) {
	addrStr := fmt.Sprintf("%s:%s", addr.IP, addr.Port)

	sock, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, sock)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func HandleSignal(fd *net.TCPListener, f *Flag) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("") // new line after Ctrl-C
		log.Println("Received SIGTERM. Server shutting down...")
		f.Exit = true
		err := fd.Close()
		if err != nil {
			log.Println("Failed to close socket")
		}

		os.Exit(0)
	}()
}

func Request(conn *net.TCPConn, bufferSize int, req utils.Payload) error {
	defer conn.Close()

	fmt.Printf("Sending Request:\n\t%-10s %s\n\t%-10s %s\n", "Message:", req.Message, "Key:", req.Key)

	if err := send(conn, &req); err != nil {
		return err
	}

	response, err := receive(conn)
	if err != nil {
		return err
	}

	fmt.Printf("Encrypted Response:\n\t%-10s %s\n\t%-10s %s\n", "Message:", response.Message, "Key:", response.Key)

	decoded := vigenere.Process(response.Message, response.Key, vigenere.Decipher)

	fmt.Println("Decrypted Message:", decoded)

	return nil
}

func receive(conn *net.TCPConn) (*utils.Payload, error) {
	// buf := make([]byte, bufferSize)
	//
	// bytesRead, err := conn.Read(buf)
	// if err != nil {
	// 	return nil, fmt.Errorf("Error: read failed: %w:", err)
	// }

	buf, err := io.ReadAll(conn)
	if err != nil {
		return nil, fmt.Errorf("Error: read failed")
	}

	var msg utils.Payload
	err = json.Unmarshal(buf, &msg)
	if err != nil {
		return nil, fmt.Errorf("Error deserializing data: %w", err)
	}
	return &msg, nil
}

func send(conn *net.TCPConn, payload *utils.Payload) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Error serializing data: %w", err)
	}

	bytesRead, err := conn.Write(response)
	if err != nil {
		return fmt.Errorf("Error write: %w", err)
	}

	if bytesRead != len(response) {
		return fmt.Errorf("Error Bytes Written: %d does not match length of message: %d\n", bytesRead, len(response))
	}

	return nil
}
