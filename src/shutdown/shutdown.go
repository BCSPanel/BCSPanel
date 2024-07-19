package shutdown

import (
	"os"

	"github.com/bddjr/BCSPanel/src/cmdserver/sharecmdlistener"
	"github.com/bddjr/BCSPanel/src/httpserver"
	"github.com/bddjr/BCSPanel/src/mylog"
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

	httpserver.CloseServerAll()
	sharecmdlistener.Close()
}
