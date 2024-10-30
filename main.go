package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bddjr/BCSPanel/src/cmdserver"
	"github.com/bddjr/BCSPanel/src/conf"
	"github.com/bddjr/BCSPanel/src/httpserver"
	"github.com/bddjr/BCSPanel/src/mygc"
	"github.com/bddjr/BCSPanel/src/myinit"
	"github.com/bddjr/BCSPanel/src/mylog"
	"github.com/bddjr/BCSPanel/src/myrand"
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

	signalCtrlC := make(chan os.Signal, 1)
	signal.Notify(signalCtrlC,
		// CTRL+C
		syscall.SIGINT,
		// kill
		syscall.SIGTERM,

		syscall.SIGHUP,
		// syscall.SIGTSTP,
	)

	myinit.Init()
	mylog.INFOln("main test RandBytes")
	myrand.RandBytes(1)

	conf.UpdateConfig()
	httpserver.Start()
	cmdserver.Start()
	go mygc.GC_laterSecond(1)

	// 捕捉停止信号
	<-signalCtrlC
	shutdown.Shutdown(1)
}
