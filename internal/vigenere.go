package internal

import (
	"log"
	"unicode"
)

// Holds the message to be ciphered/deciphered w/ key
type Msg struct {
	Content string `json:"content"`
	Key     string `json:"key"`
}

func cipher(c rune, shift rune) string {
	if c > 'a' && c < 'z' {
		c += shift % 32
	} else {

	}
}

func decipher(c rune, shift rune) string {

}

func Process(m Msg, task string) string {

	// Keep appending the key to itself until it is at least the length of the content
	for len(m.Key) < len(m.Content) {
		m.Key += m.Key
	}

	keyIndex := 0
	output := ""

	for _, c := range(m.Content) {
		if !unicode.IsLetter(c) {
			output += string(c)
			continue
		}
		shift := rune(m.Key[keyIndex])
		switch task {
		case "cipher":
			output += cipher(c, shift)
		case "decipher":
			output += decipher(c, shift)
		default: // should never reach here
			log.Fatalf("Invalid task: %s", task)
		}
		keyIndex++
	}
	return output
}
