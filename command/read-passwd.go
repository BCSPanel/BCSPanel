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
	return string(passwd), nil
}
