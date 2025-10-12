package vigenere

import (
	"log"
	"unicode"
)

type VigenereTask int

const (
	Cipher VigenereTask = iota
	Decipher
)

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

func Process(msg, key string, task VigenereTask) string {
	// Keep appending the key to itself until it is at least the length of the content
	for len(key) < len(msg) {
		key += key
	}

	keyIndex := 0
	output := ""

	for _, c := range msg {
		if !unicode.IsLetter(c) {
			output += string(c)
			continue
		}
		shift := rune(key[keyIndex])
		switch task {
		case Cipher:
			output += cipher(c, shift)
		case Decipher:
			output += decipher(c, shift)
		default: // should never reach here
			log.Fatalf("Invalid task: %v", task)
		}
		keyIndex++
	}
	return output
}
