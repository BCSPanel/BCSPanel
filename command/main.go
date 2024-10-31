package main

import (
	"fmt"
	"os"

	"github.com/bddjr/BCSPanel/src/bcspcp"
)

func errExit(err error) {
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

	wMsg := bcspcp.Message{
		"type": "command",
		"name": os.Args[1],
	}
	ctx.WriteMsg(wMsg)

	rMsg, err := ctx.ReadMsg()

	if err == nil {
		fmt.Println(rMsg)
	}
}
