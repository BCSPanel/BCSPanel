package sharecmdlistener

import (
	"net"

	"github.com/bddjr/BCSPanel/src/conf"
	"github.com/bddjr/BCSPanel/src/mylog"
)

var Listener *net.Listener = nil

var Reloading bool

func close() {
	if Listener != nil {
		if l := (*Listener); l != nil {
			l.Close()
		}
	}
}

func Reload() {
	mylog.INFOln("cmdserver Reload")
	if conf.Cmdport.OldPort != conf.Cmdport.NewPort {
		Reloading = true
		close()
	}
}

func Close() {
	Reloading = false
	mylog.INFOln("cmdserver Close")

	close()
}
