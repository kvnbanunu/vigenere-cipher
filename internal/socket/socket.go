package socket

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"vigenere-cipher/internal/utils"
	"vigenere-cipher/internal/vigenere"
)

type SockAddr struct {
	Addr *utils.Addr
}

type Flag struct {
	Exit bool
}

func (a *SockAddr) ServerSetup() (*net.TCPListener, error) {
	addrStr := a.Addr.IP + ":" + a.Addr.Port

	addr, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		return nil, err
	}

	fd, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	fmt.Println("Now listening on ", addrStr)

	return fd, nil
}

func (a *SockAddr) ClientSetup() (*net.TCPConn, error) {
	addrStr := a.Addr.IP + ":" + a.Addr.Port

	addr, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
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

func HandleConnection(conn *net.TCPConn, bufferSize int) {
	defer conn.Close()

	fmt.Println("\nConnection Accepted.")

	req, err := sock_read(conn, bufferSize)
	if err != nil {
		fmt.Println("Error reading from socket:",err)
		return
	}

	fmt.Printf("Received Payload:\n\t%-10s %s\n\t%-10s %s\n", "Message:", req.Message, "Key:", req.Key)

	// apply cipher
	encrypted := vigenere.Process(req.Message, req.Key, vigenere.Cipher)

	fmt.Println("Sending Encrypted Message:", encrypted)

	res := utils.Payload{Message: encrypted, Key: req.Key}
	if err := sock_write(conn, &res); err != nil {
		fmt.Println("Error writing to socket:",err)
		return
	}

	fmt.Println("Connection Closed.")
}

func Request(conn *net.TCPConn, bufferSize int, req utils.Payload) error {
	defer conn.Close()

	fmt.Printf("Sending Request:\n\t%-10s %s\n\t%-10s %s\n", "Message:", req.Message, "Key:", req.Key)

	if err := sock_write(conn, &req); err != nil {
		return err
	}
	
	response, err := sock_read(conn, bufferSize)
	if err != nil {
		return err
	}

	fmt.Printf("Encrypted Response:\n\t%-10s %s\n\t%-10s %s\n", "Message:", response.Message, "Key:", response.Key)

	decoded := vigenere.Process(response.Message, response.Key, vigenere.Decipher)

	fmt.Println("Decrypted Message:", decoded)

	return nil
}

func sock_read(conn *net.TCPConn, bufferSize int) (*utils.Payload, error) {
	buf := make([]byte, bufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("Error read: %w:", err)
	}

	var msg utils.Payload
	err = json.Unmarshal(buf[:n], &msg)
	if err != nil {
		return nil, fmt.Errorf("Error deserializing data: %w", err)
	}
	return &msg, nil
}

func sock_write(conn *net.TCPConn, payload *utils.Payload) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Error serializing data: %w", err)
	}

	n, err := conn.Write(response)
	if err != nil {
		return fmt.Errorf("Error write: %w", err)
	}

	if n != len(response) {
		return fmt.Errorf("Error Bytes Written: %d does not match length of message: %d\n", n, len(response))
	}

	return nil
}
