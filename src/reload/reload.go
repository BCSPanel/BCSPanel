package reload

import (
	"github.com/bddjr/BCSPanel/src/cmdserver/sharecmdlistener"
	"github.com/bddjr/BCSPanel/src/conf"
	"github.com/bddjr/BCSPanel/src/httpserver"
	"github.com/bddjr/BCSPanel/src/mygc"
	"github.com/bddjr/BCSPanel/src/mylog"
	"github.com/bddjr/BCSPanel/src/mysession"
	"github.com/bddjr/BCSPanel/src/user"
)

var Reloading bool

func Reload() {
	if Reloading {
		return
	}
	Reloading = true
	defer func() {
		Reloading = false
	}()
	mylog.INFOln("Reload")
	mysession.Reload()
	user.Reload()
	conf.UpdateConfig()
	httpserver.Reload()
	sharecmdlistener.Reload()
	go mygc.GC_laterSecond(1)
}
