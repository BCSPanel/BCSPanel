package mygc

import (
	"runtime"
	"time"

	"github.com/bddjr/BCSPanel/src/mysession"
)

var RunningGC bool

func Init() {
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			GC()
		}
	}()
}

func GC() {
	if RunningGC {
		return
	}
	RunningGC = true
	defer func() {
		RunningGC = false
	}()
	mysession.GC()
	runtime.GC()
}

func GC_later(t time.Duration) {
	time.Sleep(t)
	GC()
}

func GC_laterSecond(s int64) {
	GC_later(time.Duration(s) * time.Second)
}
