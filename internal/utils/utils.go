package utils

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
	"unicode"
)

// Holds settings for defaults
type Config struct {
	BufferSize int    `json:"bufferSize"`
	Message    string `json:"message"`
	Key        string `json:"key"`
	IP         string `json:"ip"`
	Port       string `json:"port"`
	MinDelay   uint   `json:"minDelay"`
	MaxDelay   uint   `json:"maxDelay"`
}

// Holds network socket settings
type Addr struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

// Holds the message to be ciphered/deciphered w/ key
type Payload struct {
	Message string `json:"message"`
	Key     string `json:"key"`
}

type Delay struct {
	Min uint `json:"minDelay"`
	Max uint `json:"maxDelay"`
}

// Read contents of "config.json" and store in Config struct
func LoadConfig() (*Config, error) {
	file, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Parse command line args for server
func ServerParseArgs(cfg *Config) (*Addr, *Delay) {
	prog := os.Args[0] // program name

	var help bool
	var minDelay uint
	var maxDelay uint

	flag.BoolVar(&help, "h", false, "Prints a help message")
	flag.UintVar(&minDelay, "m", cfg.MinDelay, "Minimum simulated server processing time")
	flag.UintVar(&maxDelay, "M", cfg.MaxDelay, "Maximum simulated server processing time")

	flag.Parse()
	args := flag.Args()

	if help {
		serverUsage(prog, "")
	}

	if maxDelay == minDelay {
		maxDelay++
	}

	delay := Delay{
		Min: minDelay,
		Max: maxDelay,
	}

	var addr Addr
	cfg.serverHandleArgs(prog, args, &addr)

	return &addr, &delay
}

// Parse command line args for client
func ClientParseArgs(cfg *Config) (*Payload, *Addr) {
	prog := os.Args[0]

	var help bool
	flag.BoolVar(&help, "h", false, "Prints a help message")
	flag.Parse()
	args := flag.Args()

	if help {
		clientUsage(prog, "")
	}

	var msg Payload
	var addr Addr
	cfg.clientHandleArgs(prog, args, &msg, &addr)

	return &msg, &addr
}

// Checks args for IP address and port, designed to be in any order
func (cfg *Config) serverHandleArgs(prog string, args []string, addr *Addr) {
	// insert defaults
	addr.IP = cfg.IP
	addr.Port = cfg.Port

	for i, val := range args {
		switch i {
		case 0:
			if !checkIP(val) {
				serverUsage(prog, fmt.Sprintf("Invalid IP Address: %s", val))
			}
			addr.IP = val
		case 1:
			if !checkPort(val) {
				serverUsage(prog, fmt.Sprintf("Invalid Port: %s", val))
			}
			addr.Port = val
		default:
			serverUsage(prog, "Too many arguments")
		}
	}
}

// Check for valid args, strict order
func (cfg *Config) clientHandleArgs(prog string, args []string, msg *Payload, addr *Addr) {
	// insert defaults
	msg.Message = cfg.Message
	msg.Key = cfg.Key
	addr.IP = cfg.IP
	addr.Port = cfg.Port

	for i, val := range args {
		switch i {
		case 0:
			msg.Message = val
		case 1:
			if checkKey(val) {
				msg.Key = val
			} else {
				clientUsage(prog, fmt.Sprintf("Invalid Key: %s", val))
			}
		case 2:
			if !checkIP(val) {
				clientUsage(prog, fmt.Sprintf("Invalid IP Address: %s", val))
			}
			addr.IP = val
		case 3:
			if !checkPort(val) {
				clientUsage(prog, fmt.Sprintf("Invalid Port: %s", val))
			}
			addr.Port = val
		default:
			clientUsage(prog, "Too many arguments")
		}
	}
}

func (d *Delay) SimulateDelay() {
	delay := rand.Intn(int(d.Max)-int(d.Min)) + int(d.Min)
	time.Sleep(time.Duration(delay) * time.Second)
}

// Checks each character if it is a letter in the alphabet
func checkKey(str string) bool {
	for _, c := range str {
		if !unicode.IsLetter(c) {
			return false
		}
	}
	return true
}

// Checks if the IP is a valid IP4 or IP6 address
func checkIP(str string) bool {
	ip := net.ParseIP(str)
	if ip == nil {
		return false
	}
	return true
}

// Checks if the port is valid and returns the port
func checkPort(str string) bool {
	port, err := strconv.Atoi(str)
	if err != nil {
		return false
	}

	if port < 0 || port > 65535 { // max port value
		return false
	}

	return true
}

func serverUsage(prog_name string, msg string) {
	if msg != "" {
		log.Println(msg)
	}

	str := `Usage: %s [-h] [-m] [-M] <ip address> <port>
Options:
	-h           Display this help message
	-m           Minimum simulated server processing time
	-M           Maximum simulated server processing time
	<ip address> IPv4 or IPv6 address of host
	<port>       Port to listen on
`

	fmt.Printf(str, prog_name)
	os.Exit(0)
}

func clientUsage(prog_name string, msg string) {
	if msg != "" {
		log.Println(msg)
	}

	str := `Usage: %s [-h] <msg> <key> <ip address> <port>
Options:
	-h           Display this help message
	<msg>        Message string to send
	<key>        Encryption key (Must be string of only letters)
	<ip address> IPv4 or IPv6 address of host
	<port>       Port to listen on
`

	fmt.Printf(str, prog_name)
	os.Exit(0)
}
