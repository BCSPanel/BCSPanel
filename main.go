package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bddjr/BCSPanel/src/cmdserver"
	"github.com/bddjr/BCSPanel/src/config"
	"github.com/bddjr/BCSPanel/src/httpserver"
	"github.com/bddjr/BCSPanel/src/mylog"
	"github.com/bddjr/BCSPanel/src/shutdown"
)

func main() {
	// 捕捉异常后停止
	defer func() {
		if r := recover(); r != nil {
			mylog.ERRORln(r)
			shutdown.Shutdown(1)
		}
	}()

	config.Update()
	cmdserver.Start()
	httpserver.Start()

	// 捕捉停止信号
	signalStop := make(chan os.Signal, 1)
	signal.Notify(signalStop,
		syscall.SIGINT,  // CTRL+C
		syscall.SIGTERM, // kill
		syscall.SIGHUP,
	)
	<-signalStop
	shutdown.Shutdown(1)
}
