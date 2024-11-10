package httpserver

import (
	"net"
	"net/http"
	"time"

	"github.com/bddjr/BCSPanel/src/config"
	"github.com/bddjr/BCSPanel/src/mylog"
	"github.com/bddjr/hlfhr"
)

// 80端口重定向服务，包括tcp监听
var server80 *http.Server
var server80Listening bool

// 启动80端口重定向服务。
// 该函数会阻塞运行。
func serve80() {
	if server80Listening ||
		!config.NewHttp.SSL.Enable ||
		!config.NewHttp.SSL.Listen80Port ||
		config.NewHttp.AddrPort != "443" {

		return
	}
	server80Listening = true
	defer func() {
		server80Listening = false
		server80 = nil
	}()

	mylog.INFOln("http serve80")
	server80 = &http.Server{
		Addr: net.JoinHostPort(config.NewHttp.AddrHost, "80"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hlfhr.RedirectToHttps(w, r, 302)
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}
	server80.SetKeepAlivesEnabled(false)

	err := server80.ListenAndServe()
	if err != http.ErrServerClosed {
		mylog.ERRORln(err)
	}
}

// 强制停止80端口重定向http服务，包括tcp监听
func close80() {
	if !server80Listening || server80 == nil {
		return
	}
	err := server80.Close()
	if err != nil {
		mylog.ERRORln(err)
	}
}
