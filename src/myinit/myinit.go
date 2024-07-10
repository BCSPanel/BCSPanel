package myinit

import (
	"github.com/bddjr/BCSPanel/src/mygc"
	"github.com/bddjr/BCSPanel/src/mylog"
)

// 手动调用所有Init
func Init() {
	mylog.Init()
	mygc.Init()
}
