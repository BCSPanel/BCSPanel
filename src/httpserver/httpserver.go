package httpserver

import (
	"net"
	"net/http"
	"time"

	"github.com/bddjr/BCSPanel/src/config"
	"github.com/bddjr/BCSPanel/src/httpserver/router"
	"github.com/bddjr/BCSPanel/src/mylog"
	"github.com/bddjr/hlfhr"
)

// http服务，包括tcp监听
var ServerHttp *hlfhr.Server
var ServerHttpListening bool

// 80端口重定向服务，包括tcp监听
var Server80 *http.Server
var Server80Listening bool

func Start() {
	config.SetHttpOld()
	go serverHttpListen()
}

func needRestartHttpServer() bool {
	// 当监听端口变了
	b := config.NewHttp.Addr != config.OldHttp.Addr ||
		// 改变了ssl开启状态
		config.NewHttp.SSL.Enable != config.OldHttp.SSL.Enable ||
		// 改变了H2C开启状态
		(!config.NewHttp.SSL.Enable && config.NewHttp.H2C != config.OldHttp.H2C) ||
		// 更改了路径开头
		config.NewHttp.PathPrefix != config.OldHttp.PathPrefix

	return b
}

func Reload() {
	mylog.INFOln("httpserver Reload")

	// 当改变了是否监听80端口重定向，检查是否已监听
	if Server80Listening {
		// 已监听，不符合条件就停止
		if !config.NewHttp.SSL.Enable ||
			!config.NewHttp.SSL.Listen80Port ||
			config.NewHttp.AddrPort != "443" {
			CloseServer80()
		}
	} else if config.NewHttp.SSL.Listen80Port {
		// 未监听，如果设为true就开始监听
		go server80Listen()
	}

	if needRestartHttpServer() {
		// 那么
		// 重启 ServerHttp
		mylog.INFOln("http Reload ServerHttp")
		CloseServerHttp()
		config.SetHttpOld()
		go serverHttpListen()
	} else {
		config.SetHttpOld()
	}
}

// 启动服务，包括tcp监听。
// 会自动异步启动80端口重定向服务 Server80Listen 。
// 该函数会阻塞运行。
func serverHttpListen() {
	if ServerHttpListening {
		return
	}
	ServerHttpListening = true

	{
		// print
		schemeName := "http"
		if config.OldHttp.SSL.Enable {
			schemeName += "s"
		}
		mylog.INFO(schemeName, "://localhost:", config.OldHttp.AddrPort, config.OldHttp.PathPrefix)
	}

	ServerHttp = hlfhr.New(&http.Server{
		Addr:              config.OldHttp.Addr,
		Handler:           router.GetHandler(),
		ReadHeaderTimeout: 10 * time.Second,
	})
	var err error
	if config.OldHttp.SSL.Enable {
		// https
		go server80Listen()
		err = ServerHttp.ListenAndServeTLS("", "")
	} else {
		// http
		err = ServerHttp.ListenAndServe()
	}
	if err == http.ErrServerClosed {
		// 这表明有另外的函数正在处理
		return
	}
	mylog.ERRORln(err)
	ServerHttp = nil
	ServerHttpListening = false
}

// 启动80端口重定向服务。
// 该函数会阻塞运行。
func server80Listen() {
	if Server80Listening || !config.NewHttp.SSL.Enable || !config.NewHttp.SSL.Listen80Port || config.NewHttp.AddrPort != "443" {
		return
	}
	Server80Listening = true
	mylog.INFOln("http Start Server80Listen")
	Server80 = &http.Server{
		Addr:              net.JoinHostPort(config.NewHttp.AddrHost, "80"),
		Handler:           Server80Handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	Server80.SetKeepAlivesEnabled(false)
	err := Server80.ListenAndServe()
	if err == http.ErrServerClosed {
		// 这表明有另外的函数正在处理
		return
	}
	mylog.ERRORln(err)
	Server80 = nil
	Server80Listening = false
}

// 强制停止所有服务
func CloseServerAll() {
	CloseServer80()
	CloseServerHttp()
}

// 强制停止80端口重定向http服务，包括tcp监听
func CloseServer80() {
	if !Server80Listening {
		return
	}
	mylog.INFOln("http CloseServer80")
	err := Server80.Close()
	if err != nil {
		mylog.ERRORln(err)
	}
	Server80 = nil
	Server80Listening = false
}

// 强制停止http服务，不包括tcp监听
func CloseServerHttp() {
	if !ServerHttpListening {
		return
	}
	mylog.INFOln("http CloseServerHttp")
	err := ServerHttp.Close()
	if err != nil {
		mylog.ERRORln(err)
	}
	ServerHttp = nil
	ServerHttpListening = false
}
