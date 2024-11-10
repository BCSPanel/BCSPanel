package shutdown

import (
	"os"

	"github.com/bddjr/BCSPanel/src/cmdserver/cmdservercloser"
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

	mylog.INFOln("Shutdown")

	httpserver.Close()
	cmdservercloser.Close()
}
