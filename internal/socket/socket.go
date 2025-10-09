package socket

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"vigenere-cipher/internal/utils"
	"vigenere-cipher/internal/vigenere"
)

type Server struct {
	epollFd  int
	listener int
	conns    map[int]net.Conn
}

type Flag struct {
	Exit bool
}

func serverSetup(addr *utils.Addr) (int, error) {
	var fd int
	var err error

	if addr.Type == utils.IPv4 {
		fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	} else {
		fd, err = syscall.Socket(syscall.AF_INET6, syscall.SOCK_STREAM, 0)
	}
	if err != nil {
		return 0, fmt.Errorf("Error: socket creation failed: %w", err)
	}

	// socket options
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		syscall.Close(fd)
		return 0, fmt.Errorf("Error: setsockopt failed: %w", err)
	}

	if err := syscall.SetNonblock(fd, true); err != nil {
		syscall.Close(fd)
		return 0, fmt.Errorf("Error: set nonblock failed: %w", err)
	}

	// fill sockaddr
	var sockaddr syscall.Sockaddr
	if addr.Type == utils.IPv4 {
		sockaddr = &syscall.SockaddrInet4{
			Port: addr.Port,
			Addr: [4]byte(addr.IP.To4()),
		}
	} else {
		sockaddr = &syscall.SockaddrInet6{
			Port: addr.Port,
			Addr: [16]byte(addr.IP.To16()),
		}
	}

	if err := syscall.Bind(fd, sockaddr); err != nil {
		syscall.Close(fd)
		return 0, fmt.Errorf("Error: binding failed: %w", err)
	}

	if err := syscall.Listen(fd, syscall.SOMAXCONN); err != nil {
		syscall.Close(fd)
		return 0, fmt.Errorf("Error: listen failed: %w", err)
	}

	fmt.Printf("Now listening on %s:%d\n", addr.IP.String(), addr.Port)
	return fd, nil
}

func NewServer(addr *utils.Addr) (*Server, error) {
	fd, err := serverSetup(addr)
	if err != nil {
		return nil, fmt.Errorf("Error: setting up server failed: %w", err)
	}

	// create epoll
	epollFd, err := syscall.EpollCreate1(0)
	if err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("Error: epoll_create1 failed: %w", err)
	}

	// setup epoll listener
	event := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(fd),
	}
	if err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, fd, &event); err != nil {
		syscall.Close(fd)
		syscall.Close(epollFd)
		return nil, fmt.Errorf("Error: epoll_ctl add failed: %w", err)
	}

	return &Server{
		epollFd:  epollFd,
		listener: fd,
		conns:    make(map[int]net.Conn),
	}, nil
}

func (s *Server) acceptConnection() error {
	nfd, sa, err := syscall.Accept(s.listener)
	if err != nil {
		if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
			return nil // no more connections to accept
		}
		return fmt.Errorf("Error: accept failed: %w", err)
	}
	if err := syscall.SetNonblock(nfd, true); err != nil {
		syscall.Close(nfd)
		return fmt.Errorf("Error: set nonblock failed: %w", err)
	}

	// add to epoll
	epollflags := syscall.EPOLLIN | syscall.EPOLLET
	event := syscall.EpollEvent{
		Events: uint32(epollflags),
		Fd:     int32(nfd),
	}
	if err := syscall.EpollCtl(s.epollFd, syscall.EPOLL_CTL_ADD, nfd, &event); err != nil {
		syscall.Close(nfd)
		return fmt.Errorf("Error: epoll_ctl add failed: %w", err)
	}

	// wrap the fd into a net Conn
	file := os.NewFile(uintptr(nfd), "client")
	conn, err := net.FileConn(file)
	file.Close()
	if err != nil {
		syscall.Close(nfd)
		return fmt.Errorf("Error: FileConn failed: %w", err)
	}

	s.conns[nfd] = conn

	var addr string
	switch v := sa.(type) {
	case *syscall.SockaddrInet4:
		addr = fmt.Sprintf("%d.%d.%d.%d:%d", v.Addr[0], v.Addr[1], v.Addr[2], v.Addr[3], v.Port)
	case *syscall.SockaddrInet6:
		addr = fmt.Sprintf("%s:%d", ipv6ToString(v.Addr), v.Port)
	}

	log.Printf("New connection: %s (fd: %d, total: %d)", addr, nfd, len(s.conns))
	return nil
}

func (s *Server) handleClient(fd, bufferSize int, delay *utils.Delay) error {
	fmt.Println("Now serving Client fd:", fd)
	for {
		req, err := receive(nil, fd, bufferSize)
		if err != nil {
			if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
				return nil
			}
			return err
		}
		fmt.Printf("(Client fd: %d)Received Payload:\n\t%-10s %s\n\t%-10s %s\n", fd, "Message:", req.Message, "Key:", req.Key)

		delay.SimulateDelay()
		encrypted := vigenere.Process(req.Message, req.Key, vigenere.Cipher)
		fmt.Printf("(Client fd: %d)Sending Encrypted Message: %s\n", fd, encrypted)

		res := utils.Payload{Message: encrypted, Key: req.Key}
		if err := send(nil, fd, &res); err != nil {
			return err
		}
	}
}

func (s *Server) closeConnection(fd int) {
	if conn, ok := s.conns[fd]; ok {
		conn.Close()
		delete(s.conns, fd)
	}
	syscall.EpollCtl(s.epollFd, syscall.EPOLL_CTL_DEL, fd, nil)
	syscall.Close(fd)
	log.Printf("Client fd: %d disconnected (remaining: %d)", fd, len(s.conns))
}

func (s *Server) Cleanup() {
	for fd := range s.conns {
		s.closeConnection(fd)
	}
	syscall.Close(s.listener)
	syscall.Close(s.epollFd)
}

func (s *Server) Run(bufferSize int, delay *utils.Delay) error {
	events := make([]syscall.EpollEvent, 128)

	for {
		n, err := syscall.EpollWait(s.epollFd, events, -1)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			return fmt.Errorf("Error: epoll_wait faied: %w", err)
		}

		for i := range n {
			fd := int(events[i].Fd)

			if fd == s.listener { // new connection
				if err := s.acceptConnection(); err != nil {
					log.Printf("Accept error: %v", err)
				}
			} else { // handle connection
				if events[i].Events&(syscall.EPOLLIN|syscall.EPOLLHUP|syscall.EPOLLERR) != 0 {
					if err := s.handleClient(fd, bufferSize, delay); err != nil {
						s.closeConnection(fd)
					}
				}
			}
		}
	}
}

func ipv6ToString(a [16]byte) string {
	result := []string{}
	for i := 0; i < 16; i += 2 {
		result = append(result, hex.EncodeToString(a[i:i+2]))
	}
	return strings.Join(result, ":")
}

func ClientSetup(addr *utils.Addr) (*net.TCPConn, error) {
	addrStr := fmt.Sprintf("%s:%d", addr.IP.String(), addr.Port)

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

	if err := send(conn, 0, &req); err != nil {
		return err
	}

	response, err := receive(conn, 0, bufferSize)
	if err != nil {
		return err
	}

	fmt.Printf("Encrypted Response:\n\t%-10s %s\n\t%-10s %s\n", "Message:", response.Message, "Key:", response.Key)

	decoded := vigenere.Process(response.Message, response.Key, vigenere.Decipher)

	fmt.Println("Decrypted Message:", decoded)

	return nil
}

func receive(conn *net.TCPConn, fd int, bufferSize int) (*utils.Payload, error) {
	buf := make([]byte, bufferSize)
	var bytesRead int
	var err error

	if conn != nil {
		bytesRead, err = conn.Read(buf)
	} else {
		bytesRead, err = syscall.Read(fd, buf)
	}

	if err != nil {
		return nil, fmt.Errorf("Error read: %w:", err)
	}

	if bytesRead == 0 {
		return nil, fmt.Errorf("Error: client closed connection")
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
		bytesRead, err = syscall.Write(fd, response)
	}

	if err != nil {
		return fmt.Errorf("Error write: %w", err)
	}

	if bytesRead != len(response) {
		return fmt.Errorf("Error Bytes Written: %d does not match length of message: %d\n", bytesRead, len(response))
	}

	return nil
}
