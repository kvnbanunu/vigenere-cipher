package internal

import (
	"encoding/json"
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
	args := os.Args

	if len(args) > 3 { // Program name, IP, Port
		serverUsage(args[0], "Too many arguments.")
	}

	if len(args) > 1 && args[1] == "-h" {
		serverUsage(args[0], "")
	}

	var addr Addr
	cfg.serverHandleArgs(args, &addr)

	return &addr
}

// Parse command line args for client
func ClientParseArgs(cfg *Config) (*Msg, *Addr) {
	args := os.Args

	if len(args) > 5 { // Program name, msg, key, IP, Port
		clientUsage(args[0], "Too many arguments.")
	}

	if len(args) > 1 && args[1] == "-h" {
		clientUsage(args[0], "")
	}

	var msg Msg
	var addr Addr
	cfg.clientHandleArgs(args, &msg, &addr)

	return &msg, &addr
}

// Checks args for IP address and port, designed to be in any order
func (cfg *Config) serverHandleArgs(args []string, addr *Addr) {
	numArgs := len(args)
	hasIP := false
	hasPort := false

	// loop over args, but skip prog_name
	for i := 1; i < numArgs; i++ {
		isIP := checkIP(args[i])

		if !hasIP && isIP {
			addr.IP = args[i]
			hasIP = true
			continue
		} else if hasIP && isIP {
			// included more than 1 ip address
			serverUsage(args[0], "Inputted too many IP addresses")
		}

		isPort := checkPort(args[i])

		if !hasPort && isPort {
			addr.Port = args[i]
			hasPort = true
			continue
		} else if hasPort && isPort {
			serverUsage(args[0], "Inputted too many Ports")
		}

		// if the arg is neither an address or port
		serverUsage(args[0], fmt.Sprintf("Invalid argument: %s", args[i]))
	}

	// Insert defaults if empty
	if !hasIP {
		addr.IP = cfg.IP
	}

	if !hasPort {
		addr.Port = cfg.Port
	}
}

// Check for valid args, strict order
func (cfg *Config) clientHandleArgs(args []string, msg *Msg, addr *Addr) {
	numArgs := len(args)

	// insert defaults
	msg.Content = cfg.Content
	msg.Key = cfg.Key
	addr.IP = cfg.IP
	addr.Port = cfg.Port

	// if no args past prog_name
	if numArgs == 1 {
		return
	}

	msg.Content = args[1]

	if numArgs == 2 {
		return
	}

	key := args[2]
	if checkKey(key) {
		msg.Key = key
	} else {
		clientUsage(args[0], fmt.Sprintf("Invalid Key: %s", key))
	}

	if numArgs == 3 {
		return
	}

	ip := args[3]
	if checkIP(ip) {
		addr.IP = ip
	} else {
		clientUsage(args[0], fmt.Sprintf("Invalid IP Address: %s", ip))
	}

	if numArgs == 4 {
		return
	}

	port := args[4]
	if checkPort(port) {
		addr.Port = port
	} else {
		clientUsage(args[0], fmt.Sprintf("Invalid Port: %s", port))
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
func checkIP(str string) bool {
	ip := strings.Split(str, ".")

	// if '.' exists in str then there should be more than 1 element
	// and an ipv4 addr should have 4 elements ex. 0.0.0.0
	if len(ip) == 4 {
		for _, segment := range ip {
			if len(segment) > 3 {
				return false
			}

			num, err := strconv.Atoi(segment)
			if err != nil {
				return false
			}

			if num > 255 {
				return false
			}
		}
		return true
	} else {
		ip6 := strings.Split(str, ":")
		if len(ip6) >= 3 { // shortest representation is "::" is still 3 elements
			for _, segment := range ip6 {
				if segment == "" { // represents a compressed 0 segment
					continue
				}

				if len(segment) > 4 { // each segment has a max length of 4 hex digits
					return false
				}

				num, err := strconv.ParseUint(segment, 16, 64)
				if err != nil {
					return false
				}

				if num > 65535 { // max int value for FFFF
					return false
				}
			}
			return true
		}
	}

	return false
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

	str := `
Usage: %s [-h] <ip address> <port>
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

	str := `
Usage: %s [-h] <msg> <key> <ip address> <port>
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
