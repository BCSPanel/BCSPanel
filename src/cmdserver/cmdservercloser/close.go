package cmdservercloser

import (
	"github.com/bddjr/BCSPanel/src/mylog"
)

var closer func() = nil

func SetCloser(f func()) {
	closer = f
}

func Close() {
	mylog.INFOln("cmdserver Close")
	if closer != nil {
		closer()
		closer = nil
	}
}
