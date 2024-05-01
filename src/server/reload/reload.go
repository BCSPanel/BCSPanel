package reload

import (
	"github.com/bddjr/BCSPanel/src/server/cmdserver/sharecmdlistener"
	"github.com/bddjr/BCSPanel/src/server/conf"
	"github.com/bddjr/BCSPanel/src/server/httpserver"
	"github.com/bddjr/BCSPanel/src/server/mygc"
	"github.com/bddjr/BCSPanel/src/server/mylog"
	"github.com/bddjr/BCSPanel/src/server/mysession"
	"github.com/bddjr/BCSPanel/src/server/user"
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
