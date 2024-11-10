package httpserver

import (
	"net/http"
	"time"

	"github.com/bddjr/BCSPanel/src/config"
	"github.com/bddjr/BCSPanel/src/httpserver/router"
	"github.com/bddjr/BCSPanel/src/mylog"
	"github.com/bddjr/hlfhr"
)

// http服务，包括tcp监听
var server *hlfhr.Server
var serverListening bool

// 启动服务，包括tcp监听。
// 会自动异步启动80端口重定向服务 Server80Listen 。
// 该函数会阻塞运行。
func serve() {
	if serverListening {
		return
	}
	serverListening = true
	defer func() {
		serverListening = false
		server = nil
	}()
	config.SetHttpOld()

	{
		// print
		schemeName := "http"
		if config.OldHttp.SSL.Enable {
			schemeName += "s"
		}
		mylog.INFO(schemeName, "://localhost:", config.OldHttp.AddrPort, config.OldHttp.PathPrefix)
	}

	server = hlfhr.New(&http.Server{
		Addr:              config.OldHttp.Addr,
		Handler:           router.GetHandler(),
		ReadHeaderTimeout: 10 * time.Second,
	})

	var err error
	if config.OldHttp.SSL.Enable {
		// https
		err = server.ListenAndServeTLS("", "")
	} else {
		// http
		err = server.ListenAndServe()
	}

	if err != http.ErrServerClosed {
		mylog.ERRORln(err)
	}
}

// 强制停止http服务，不包括tcp监听
func close() {
	if !serverListening || server == nil {
		return
	}
	err := server.Close()
	if err != nil {
		mylog.ERRORln(err)
	}
	server = nil
	serverListening = false
}
