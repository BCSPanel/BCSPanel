package shutdown

import (
	"os"
	"sync"

	"github.com/bddjr/BCSPanel/src/server/cmdserver/sharecmdlistener"
	"github.com/bddjr/BCSPanel/src/server/httpserver"
	"github.com/bddjr/BCSPanel/src/server/mylog"
)

var Shutingdown bool

func Shutdown(code int) {
	if Shutingdown {
		return
	}
	Shutingdown = true

	defer os.Exit(code)
	defer os.RemoveAll("cache")

	mylog.INFOln("Shutdown")

	var wg sync.WaitGroup
	wg.Add(1)
	go httpserver.ShutdownServerAll(&wg)
	wg.Wait()

	sharecmdlistener.Close(nil)
}
