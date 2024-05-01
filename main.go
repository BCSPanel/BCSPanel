package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bddjr/BCSPanel/src/server/cmdclient"
	"github.com/bddjr/BCSPanel/src/server/cmdserver"
	"github.com/bddjr/BCSPanel/src/server/conf"
	"github.com/bddjr/BCSPanel/src/server/httpserver"
	"github.com/bddjr/BCSPanel/src/server/isservermode"
	"github.com/bddjr/BCSPanel/src/server/mygc"
	"github.com/bddjr/BCSPanel/src/server/myinit"
	"github.com/bddjr/BCSPanel/src/server/mylog"
	"github.com/bddjr/BCSPanel/src/server/myrand"
	"github.com/bddjr/BCSPanel/src/server/shutdown"
)

func main() {
	// 捕捉异常后停止
	defer func() {
		if r := recover(); r != nil {
			mylog.ERRORln(r)
			mainShutdown(1)
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

	// 捕捉停止信号
	go func() {
		<-signalCtrlC
		mainShutdown(1)
	}()

	if isservermode.IsServerMode {
		myinit.Init()
		mylog.INFOln("main test RandBytes")
		myrand.RandBytes(1)

		conf.UpdateConfig()
		httpserver.Start()
		cmdserver.Start()
		go mygc.GC_laterSecond(1)

		// 检测reload信号
		// go func() {
		// 	for {
		// 		signalReload := make(chan os.Signal, 1)
		// 		signal.Notify(signalReload, syscall.Signal(30))
		// 		<-signalReload
		// 		reload.Reload()
		// 	}
		// }()

		// 永远阻塞
		select {}
	} else {
		cmdclient.Run()
	}
}

func mainShutdown(code int) {
	if isservermode.IsServerMode {
		shutdown.Shutdown(code)
	}
	os.Exit(code)
}
