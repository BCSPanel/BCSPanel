package main

import (
	"fmt"

	"golang.org/x/term"
)

func readPassword(print string) (string, error) {
	fmt.Print(print)
	defer fmt.Println()

	var err error
	oldState, err = term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState)

	passwd, err := term.ReadPassword(fd)
	if err != nil {
		return "", err
	}

	str := ""
	for _, b := range passwd {
		if b > 31 && b < 127 {
			// ASCII visible characters
			str += string(b)
		} else if b == 127 {
			// backspace
			if str != "" {
				str = str[:len(str)-1]
			}
		}
		// ignore other key
	}
	return str, nil
}
