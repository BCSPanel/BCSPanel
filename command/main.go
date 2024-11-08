package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"

	"github.com/bddjr/BCSPanel/src/bcspcp"
	"github.com/bddjr/BCSPanel/src/user"
)

var fd = int(os.Stdin.Fd())
var oldState *term.State

func errExit(err any) {
	fmt.Println("error:", err)
	os.Exit(1)
}

func main() {
	if len(os.Args) == 0 {
		return
	}

	ctx, err := bcspcp.Dial("")
	if err != nil {
		errExit(err)
	}
	defer ctx.Close()

	signalStop := make(chan os.Signal, 1)
	signal.Notify(signalStop,
		syscall.SIGINT,  // CTRL+C
		syscall.SIGTERM, // kill
		syscall.SIGHUP,
	)
	go func() {
		<-signalStop
		term.Restore(fd, oldState)
		os.Exit(1)
	}()

	name := os.Args[1]
	wMsg := bcspcp.Message{
		"type": "command",
		"name": name,
	}
	switch name {
	case "reload", "shutdown":
		//

	case "register":
		var username string
		if len(os.Args) > 2 {
			username = os.Args[2]
		} else {
			fmt.Print("User name: ")
			_, err := fmt.Scanln(&username)
			if err != nil {
				errExit(err)
			}
		}

		if !user.RegexpUsernameFormat.MatchString(username) {
			errExit("name does not conform format: " + `^[\w\-]{1,32}$`)
		}

		var passwd string
		if len(os.Args) > 3 {
			passwd = os.Args[3]
			if len(passwd) < 10 {
				errExit("password length < 10")
			}
		} else {
			passwd, err = readPassword("Password: ")
			if err != nil {
				errExit(err)
			}
			repeatPasswd, err := readPassword("Repeat password: ")
			fmt.Println()
			if err != nil {
				errExit(err)
			}
			if len(passwd) < 10 {
				errExit("password length < 10")
			}
			if passwd != repeatPasswd {
				errExit("password and repeat password does not match")
			}
		}

		wMsg["username"] = username
		wMsg["password"] = passwd

	default:
		errExit("unknown command")
	}
	ctx.WriteMsg(wMsg)

	rMsg, err := ctx.ReadMsg()
	if err != nil {
		if name == "shutdown" {
			return
		}
		errExit(err)
	}

	if err, ok := rMsg["error"]; ok {
		errExit(err)
	}
}
