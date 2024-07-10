package myinit

import (
	"github.com/bddjr/BCSPanel/src/server/mygc"
	"github.com/bddjr/BCSPanel/src/server/mylog"
)

// 手动调用所有Init
func Init() {
	mylog.Init()
	mygc.Init()
}
