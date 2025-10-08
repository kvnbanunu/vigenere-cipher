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

	buf := make([]byte, bufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error read:", err)
		return
	}

	var msg utils.Payload
	err = json.Unmarshal(buf[:n], &msg)
	if err != nil {
		fmt.Println("Error deserializing data:", err)
		return
	}

	fmt.Printf("Received Payload:\n\t%-10s %s\n\t%-10s %s\n", "Message:", msg.Message, "Key:", msg.Key)

	// apply cipher
	msg.Message = vigenere.Process(msg.Message, msg.Key, vigenere.Cipher)

	fmt.Println("Sending Encrypted Message:", msg.Message)

	response, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Error serializing data:", err)
		return
	}

	n, err = conn.Write(response)
	if err != nil {
		fmt.Println("Error write:", err)
		return
	}

	if n != len(response) {
		fmt.Printf("Error Bytes Written: %d does not match length of message: %d\n", n, len(response))
		return
	}

	fmt.Println("Connection Closed.")
}

func Request(conn *net.TCPConn, bufferSize int, msg utils.Payload) error {
	defer conn.Close()

	encoded, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	fmt.Printf("Sending Request:\n\t%-10s %s\n\t%-10s %s\n", "Message:", msg.Message, "Key:", msg.Key)

	n, err := conn.Write(encoded)
	if err != nil {
		return err
	}

	if n != len(encoded) {
		return fmt.Errorf("Bytes Written: %d does not match length of message: %d\n", n, len(encoded))
	}

	buf := make([]byte, bufferSize)
	n, err = conn.Read(buf)
	if err != nil {
		return err
	}

	var response utils.Payload
	err = json.Unmarshal(buf[:n], &response)
	if err != nil {
		return err
	}

	fmt.Printf("Encrypted Response:\n\t%-10s %s\n\t%-10s %s\n", "Message:", response.Message, "Key:", response.Key)

	decoded := vigenere.Process(response.Message, response.Key, vigenere.Decipher)

	fmt.Println("Decrypted Message:", decoded)

	return nil
}
