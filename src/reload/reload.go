package reload

import (
	"github.com/bddjr/BCSPanel/src/config"
	"github.com/bddjr/BCSPanel/src/httpserver"
	"github.com/bddjr/BCSPanel/src/mylog"
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
	user.Reload()
	config.Update()
	httpserver.Reload()
}
