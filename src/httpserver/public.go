package httpserver

import (
	"github.com/bddjr/BCSPanel/src/config"
	"github.com/bddjr/BCSPanel/src/mylog"
)

func Start() {
	mylog.INFO("http Start")
	go serve()
	go serve80()
}

// 强制停止所有服务
func Close() {
	mylog.INFO("http Close")
	close()
	close80()
}

func Reload() {
	mylog.INFOln("httpserver Reload")

	// 当监听端口变了
	if config.NewHttp.Addr != config.OldHttp.Addr ||
		// 改变了ssl开启状态
		config.NewHttp.SSL.Enable != config.OldHttp.SSL.Enable ||
		// 改变了H2C开启状态
		(!config.NewHttp.SSL.Enable && config.NewHttp.H2C != config.OldHttp.H2C) ||
		// 更改了路径开头
		config.NewHttp.PathPrefix != config.OldHttp.PathPrefix {

		// 重启
		close()
		go serve()
	} else {
		config.SetHttpOld()
	}

	// 重启80
	close80()
	go serve80()
}
