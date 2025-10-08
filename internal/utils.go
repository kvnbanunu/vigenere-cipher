package internal

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// Holds settings for defaults
type Config struct {
	BufferSize int    `json:"bufferSize"`
	Content    string `json:"content"`
	Key        string `json:"key"`
	Type       IPType `json:"ipType"`
	IP         string `json:"ip"`
	Port       string `json:"port"`
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
func ServerParseArgs(cfg *Config) *Addr {
	prog := os.Args[0] // program name

	var help bool
	flag.BoolVar(&help, "h", false, "Prints a help message")
	flag.Parse()
	args := flag.Args()

	if help {
		serverUsage(prog, "")
	}

	if len(args) > 3 { // Program name, IP, Port
		serverUsage(prog, "Too many arguments.")
	}

	var addr Addr
	cfg.serverHandleArgs(prog, args, &addr)

	return &addr
}

// Parse command line args for client
func ClientParseArgs(cfg *Config) (*Msg, *Addr) {
	prog := os.Args[0]

	var help bool
	flag.BoolVar(&help, "h", false, "Prints a help message")
	flag.Parse()
	args := flag.Args()

	if help {
		clientUsage(prog, "")
	}

	if len(args) > 5 { // Program name, msg, key, IP, Port
		clientUsage(prog, "Too many arguments.")
	}

	var msg Msg
	var addr Addr
	cfg.clientHandleArgs(prog, args, &msg, &addr)

	return &msg, &addr
}

// Checks args for IP address and port, designed to be in any order
func (cfg *Config) serverHandleArgs(prog string, args []string, addr *Addr) {
	// insert defaults
	addr.Type = cfg.Type
	addr.IP = cfg.IP
	addr.Port = cfg.Port

	for i, val := range args {
		switch i {
		case 0:
			ipType := checkIP(val)
			if ipType == BadIP {
				serverUsage(prog, fmt.Sprintf("Invalid IP Address: %s", val))
			}
			addr.Type = ipType
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
func (cfg *Config) clientHandleArgs(prog string, args []string, msg *Msg, addr *Addr) {
	// insert defaults
	msg.Content = cfg.Content
	msg.Key = cfg.Key
	addr.Type = cfg.Type
	addr.IP = cfg.IP
	addr.Port = cfg.Port

	for i, val := range args {
		switch i {
		case 0:
			msg.Content = val
		case 1:
			if checkKey(val) {
				msg.Key = val
			} else {
				clientUsage(prog, fmt.Sprintf("Invalid Key: %s", val))
			}
		case 2:
			ipType := checkIP(val)
			if ipType == BadIP {
				clientUsage(prog, fmt.Sprintf("Invalid IP Address: %s", val))
			}
			addr.Type = ipType
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
func checkIP(str string) IPType {
	ip := strings.Split(str, ".")

	// if '.' exists in str then there should be more than 1 element
	// and an ipv4 addr should have 4 elements ex. 0.0.0.0
	if len(ip) == 4 {
		for _, segment := range ip {
			if len(segment) > 3 {
				return BadIP
			}

			num, err := strconv.Atoi(segment)
			if err != nil {
				return BadIP
			}

			if num > 255 {
				return BadIP
			}
		}
		return IPv4
	} else {
		ip6 := strings.Split(str, ":")
		if len(ip6) >= 3 { // shortest representation is "::" is still 3 elements
			for _, segment := range ip6 {
				if segment == "" { // represents a compressed 0 segment
					continue
				}

				if len(segment) > 4 { // each segment has a max length of 4 hex digits
					return BadIP
				}

				num, err := strconv.ParseUint(segment, 16, 64)
				if err != nil {
					return BadIP
				}

				if num > 65535 { // max int value for FFFF
					return BadIP
				}
			}
			return IPv6
		}
	}

	return BadIP
}

// Checks if the port is valid
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

	str := `Usage: %s [-h] <ip address> <port>
Options:
	-h           Display this help message
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
