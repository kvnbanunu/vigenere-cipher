package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// Holds network socket settings
type Addr struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

type Flag struct {
	Exit bool
}

func (a *Addr) ServerSetup() (*net.TCPListener, error) {
	addrStr := a.IP + ":" + a.Port

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

func (a *Addr) ClientSetup() (*net.TCPConn, error) {
	addrStr := a.IP + ":" + a.Port

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

func HandleConnection(conn *net.TCPConn, bufferSize int) error {
	defer conn.Close()

	fmt.Println("Connection Accepted.")

	buf := make([]byte, bufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	var msg Msg
	err = json.Unmarshal(buf[:n], &msg)
	if err != nil {
		return err
	}

	fmt.Printf("Received Message:\n\t%-10s %s\n\t%-10s %s\n", "Content:", msg.Content, "Key:", msg.Key)

	// apply cipher
	msg.Content = Process(msg, "cipher")

	fmt.Println("Sending Encrypted Message:", msg.Content)

	response, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	n, err = conn.Write(response)
	if err != nil {
		return err
	}

	if n != len(response) {
		return fmt.Errorf("Bytes Written: %d does not match length of message: %d\n", n, len(response))
	}

	fmt.Println("Connection Closed.")

	return nil
}

func Request(conn *net.TCPConn, bufferSize int, msg Msg) error {
	defer conn.Close()

	encoded, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	fmt.Printf("Sending Request:\n\t%-10s %s\n\t%-10s %s\n", "Content:", msg.Content, "Key:", msg.Key)

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

	var response Msg
	err = json.Unmarshal(buf[:n], &response)
	if err != nil {
		return err
	}

	fmt.Printf("Encrypted Response:\n\t%-10s %s\n\t%-10s %s\n", "Content:", response.Content, "Key:", response.Key)

	decoded := Process(response, "decipher")

	fmt.Println("Decrypted Message:", decoded)

	return nil
}
