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
	shift %= 32 // converts ascii value to position in alphabet
	shift--     // shift is zero indexed

	if c >= 'a' && c <= 'z' {
		if shift > 'z'-c {
			c += shift - 26
		} else {
			c += shift
		}
	} else {
		if shift > 'Z'-c {
			c += shift - 26
		} else {
			c += shift
		}
	}
	return string(c)
}

func decipher(c rune, shift rune) string {
	shift %= 32 // converts ascii value to position in alphabet
	shift--     // shift is zero indexed

	if c >= 'a' && c <= 'z' {
		if shift > c-'a' {
			c -= shift - 26
		} else {
			c -= shift
		}
	} else {
		if shift > c-'A' {
			c -= shift - 26
		} else {
			c -= shift
		}
	}
	return string(c)
}

func Process(m Msg, task string) string {
	// Keep appending the key to itself until it is at least the length of the content
	for len(m.Key) < len(m.Content) {
		m.Key += m.Key
	}

	keyIndex := 0
	output := ""

	for _, c := range m.Content {
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
