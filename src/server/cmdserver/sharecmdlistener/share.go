package sharecmdlistener

import (
	"net"
	"sync"

	"github.com/bddjr/BCSPanel/src/server/conf"
	"github.com/bddjr/BCSPanel/src/server/mylog"
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

func Close(inwg *sync.WaitGroup) {
	Reloading = false
	defer func() {
		if inwg != nil {
			inwg.Done()
		}
	}()
	mylog.INFOln("cmdserver Close")

	close()
}
