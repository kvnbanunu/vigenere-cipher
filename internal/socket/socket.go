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

	"golang.org/x/sys/unix"
)

type Request struct {
	Id      int
	Payload *utils.Payload
}

type Server struct {
	Host      int
	Fds       []unix.PollFd
	ConnCount int
	Requests  map[int32]Request
	ExitFlag  bool
}

func ServerSetup(addr *utils.Addr) (*Server, error) {
	addrStr := addr.IP + ":" + addr.Port
	var server Server

	sock, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP("tcp", sock)
	if err != nil {
		return nil, err
	}

	fmt.Println("Now listening on ", addrStr)

	listenerFile, err := listener.File()
	if err != nil {
		listener.Close()
		return nil, err
	}

	server.Host = int(listenerFile.Fd())
	if err := unix.SetNonblock(server.Host, true); err != nil {
		listener.Close()
		return nil, err
	}

	server.Fds = []unix.PollFd{{Fd: int32(server.Host), Events: unix.POLLIN}}
	server.ConnCount = 0
	server.Requests = make(map[int32]Request)
	server.ExitFlag = false

	return &server, nil
}

func (s *Server) Run(bufferSize int, delay *utils.Delay) error {
	for !s.ExitFlag {
		n, err := unix.Poll(s.Fds, 0)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			return err
		}
		if n == 0 {
			continue
		}

		for _, val := range s.Fds {
			fd := val.Fd
			revents := val.Revents

			// new connection
			switch {
			case fd == int32(s.Host) && (revents&unix.POLLIN != 0):
				err := s.addConnection()
				if err != nil {
					return fmt.Errorf("Error adding connection: %w", err)
				}
			case revents&unix.POLLIN != 0:
				err := s.handleInput(fd, bufferSize)
				if err != nil {
					s.removeClient(fd)
				}
			case revents&unix.POLLOUT != 0:
				s.handleOutput(fd)
			}
			delay.SimulateDelay()
		}
	}
	return nil
}

func (s *Server) addConnection() error {
	client, sockaddr, err := unix.Accept(s.Host)
	if err != nil {
		if err == unix.EWOULDBLOCK {
			return nil
		}
		return err
	}
	unix.SetNonblock(client, true)
	addr := sockaddr.(*unix.SockaddrInet4)
	s.Requests[int32(client)] = Request{Id: s.ConnCount, Payload: &utils.Payload{}}
	ip := net.IPv4(addr.Addr[0], addr.Addr[1], addr.Addr[2], addr.Addr[3])
	fmt.Printf(`
[Client #%d]
	Accepted connection from %s:%d
`, s.ConnCount, ip.String(), addr.Port)
	s.ConnCount++
	s.Fds = append(s.Fds, unix.PollFd{Fd: int32(client), Events: unix.POLLIN})
	return nil
}

func (s *Server) handleInput(fd int32, bufferSize int) error {
	clientID := s.Requests[fd].Id

	req, err := receive(nil, int(fd), bufferSize)
	if err != nil {
		return fmt.Errorf("Error receiving request: %w", err)
	}

	fmt.Printf("\n[Client #%d]\n\tPayload Received\n\t%-10s%s\n\t%-10s%s\n",
		clientID, "Message:", req.Message, "Key:", req.Key)

	// set this client to pollout
	s.updateEvents(fd, unix.POLLOUT)
	s.Requests[fd].Payload.Message = req.Message
	s.Requests[fd].Payload.Key = req.Key
	return nil
}

func (s *Server) handleOutput(fd int32) {
	clientID := s.Requests[fd].Id
	req := s.Requests[fd].Payload
	encrypted := vigenere.Process(req.Message, req.Key, vigenere.Cipher)

	fmt.Printf(`
[Client #%d]
	Sending encrypted message: %s
`, clientID, encrypted)

	req.Message = encrypted
	err := send(nil, int(fd), req)
	if err != nil {
		log.Println("Error sending response:", err)
		// closing connection anyways don't return
	}

	fmt.Printf(`	Closing Client #%d connection
	`, clientID)
	s.updateEvents(fd, unix.POLLIN)
	s.removeClient(fd)
}

func (s *Server) updateEvents(fd int32, events int16) {
	for i := range s.Fds {
		if s.Fds[i].Fd == fd {
			s.Fds[i].Events = events
		}
	}
}

func (s *Server) removeClient(fd int32) {
	delete(s.Requests, fd)
	for i, f := range s.Fds {
		if f.Fd == fd {
			unix.Close(int(fd))
			s.Fds = append(s.Fds[:i], s.Fds[i+1:]...)
		}
	}
}

func (s *Server) Cleanup() {
	s.ExitFlag = true
	for _, val := range s.Fds {
		fd := int(val.Fd)
		unix.Close(fd)
	}
	unix.Close(s.Host)
	fmt.Println("Cleanup Success. Server shutting down.")
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

func HandleSignal(server *Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("") // new line after Ctrl-C
		log.Println("Received SIGTERM. Start cleanup.")
		server.Cleanup()
		os.Exit(0)
	}()
}

func SendRequest(conn *net.TCPConn, bufferSize int, req utils.Payload) error {
	defer conn.Close()

	fmt.Printf("Sending Request:\n\t%-10s %s\n\t%-10s %s\n", "Message:", req.Message, "Key:", req.Key)

	if err := send(conn, 0, &req); err != nil {
		return err
	}

	response, err := receive(conn, 0, bufferSize)
	if err != nil {
		return err
	}

	fmt.Println("Encrypted Response:\t", response.Message)

	decoded := vigenere.Process(response.Message, response.Key, vigenere.Decipher)

	fmt.Println("Decrypted Message:\t", decoded)

	return nil
}

func receive(conn *net.TCPConn, fd, bufferSize int) (*utils.Payload, error) {
	buf := make([]byte, bufferSize)
	var bytesRead int
	var err error

	if conn != nil {
		bytesRead, err = conn.Read(buf)
	} else {
		bytesRead, err = unix.Read(fd, buf)
	}

	if err != nil {
		return nil, fmt.Errorf("Error read: %w", err)
	}

	if bytesRead == 0 {
		return nil, fmt.Errorf("Client closed connection: %w", err)
	}

	var msg utils.Payload
	err = json.Unmarshal(buf[:bytesRead], &msg)
	if err != nil {
		return nil, fmt.Errorf("Error deserializing data: %w", err)
	}
	return &msg, nil
}

func send(conn *net.TCPConn, fd int, payload *utils.Payload) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Error serializing data: %w", err)
	}

	var bytesRead int

	if conn != nil {
		bytesRead, err = conn.Write(response)
	} else {
		bytesRead, err = unix.Write(fd, response)
	}
	if err != nil {
		return fmt.Errorf("Error write: %w", err)
	}

	if bytesRead != len(response) {
		return fmt.Errorf("Error Bytes Written: %d does not match length of message: %d\n", bytesRead, len(response))
	}

	return nil
}
