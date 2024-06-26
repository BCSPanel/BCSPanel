package httpserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bddjr/BCSPanel/src/server/conf"
	"github.com/bddjr/BCSPanel/src/server/httprouter"
	"github.com/bddjr/BCSPanel/src/server/mylog"
	"github.com/bddjr/hlfhr"
)

// http服务，包括tcp监听
var ServerHttp *hlfhr.Server
var ServerHttpListening bool

// 80端口重定向服务，包括tcp监听
var Server80 *http.Server
var Server80Listening bool

func Start() {
	// 向屏幕输出调试地址
	var sb strings.Builder
	sb.WriteString("\n    http")
	if conf.Ssl.New_EnableSsl {
		sb.WriteString("s")
	}
	sb.WriteString("://localhost")
	var httpDefaultPort uint16
	if conf.Ssl.New_EnableSsl {
		httpDefaultPort = 443
	} else {
		httpDefaultPort = 80
	}
	if httpDefaultPort != conf.Http.New_ServerHttpPortNumber {
		// 端口不是协议标准的
		sb.WriteString(fmt.Sprintf(":%d", conf.Http.New_ServerHttpPortNumber))
	} else if conf.Ssl.New_EnableSsl && conf.Ssl.Only_EnableListen80Redirect {
		// 端口是443，开启了https，开启了80监听重定向
		sb.WriteString("\n    http://localhost\n")
	}
	sb.WriteString(conf.Http.New_PathPrefix)
	sb.WriteString("\n")
	fmt.Println(sb.String())
	// 启动
	go ServerHttpListen()
}

func Reload() {
	mylog.INFOln("httpserver Reload")
	// 当改变了是否监听80端口重定向，检查是否已监听
	if Server80Listening {
		// 已监听，不符合条件就停止
		if !conf.Ssl.New_EnableSsl || !conf.Ssl.Only_EnableListen80Redirect || conf.Http.New_ServerHttpPortNumber != 443 {
			ShundownServer80(nil)
		} else if conf.Http.Old_Server80Port != conf.Http.New_Server80Port {
			// 监听端口变了
			ShundownServer80(nil)
			time.Sleep(1 * time.Second)
			go Server80Listen()
		}
	} else if conf.Ssl.Only_EnableListen80Redirect {
		// 未监听，如果设为true就开始监听
		go Server80Listen()
	}
	// 当监听端口变了，或，或，或，或
	if conf.Http.Old_ServerHttpPort != conf.Http.New_ServerHttpPort ||
		// 改变了ssl开启状态
		conf.Ssl.Old_EnableSsl != conf.Ssl.New_EnableSsl ||
		// 更改了保持连接时长
		conf.Http.Old_KeepAliveSecond != conf.Http.New_KeepAliveSecond ||
		// 更改了gzip压缩等级
		conf.Http.Old_GzipLevel != conf.Http.New_GzipLevel ||
		// 改变了H2C开启状态
		(!conf.Ssl.New_EnableSsl && conf.Http.Old_EnableH2c != conf.Http.New_EnableH2c) ||
		// 更改了路径开头
		conf.Http.Old_PathPrefix != conf.Http.New_PathPrefix {
		// 那么
		// 重启 ServerHttp
		mylog.INFOln("http Reload ServerHttp")
		ShundownServerHttp(nil)
		// 等1秒再启动
		time.Sleep(1 * time.Second)
		go ServerHttpListen()
		return
	}
	if conf.Ssl.New_EnableSsl {
		// 如果已开启https，检查是否启用http2
		if conf.Ssl.Only_EnableHttp2 {
			// 启用
			ServerHttp.TLSNextProto = nil
		} else if ServerHttp.TLSNextProto == nil {
			// 禁用
			ServerHttp.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}
		}
	}
	httprouter.UpdateColorScheme()
	httprouter.UPdate404html()
}

// 启动服务，包括tcp监听。
// 会自动异步启动80端口重定向服务 Server80Listen 。
// 该函数会阻塞运行。
func ServerHttpListen() {
	if ServerHttpListening {
		return
	}
	ServerHttpListening = true
	conf.Http.Old_ServerHttpPort = conf.Http.New_ServerHttpPort
	conf.Http.Old_ServerHttpPortNumber = conf.Http.New_ServerHttpPortNumber
	conf.Ssl.Old_EnableSsl = conf.Ssl.New_EnableSsl
	conf.Http.Old_EnableH2c = conf.Http.New_EnableH2c
	conf.Http.Old_KeepAliveSecond = conf.Http.New_KeepAliveSecond

	mylog.INFOf("http ServerHttpListen port %s , ssl %v\n", conf.Http.Old_ServerHttpPort, conf.Ssl.Old_EnableSsl)
	httprouter.UpdateRouter()
	httprouter.Router.UseH2C = conf.Http.Old_EnableH2c && !conf.Ssl.Old_EnableSsl // https://github.com/gin-gonic/gin/pull/1398
	ServerHttp = hlfhr.New(&http.Server{
		Addr:              conf.Http.Old_ServerHttpPort,
		Handler:           httprouter.Router.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	})
	// 是否保持连接
	if conf.Http.Old_KeepAliveSecond < 0 {
		// 否
		ServerHttp.SetKeepAlivesEnabled(false)
	} else {
		// 是
		ServerHttp.IdleTimeout = time.Duration(conf.Http.Old_KeepAliveSecond) * time.Second
	}
	var err error
	if !conf.Ssl.Old_EnableSsl {
		// http
		err = ServerHttp.ListenAndServe()
	} else {
		// https
		go Server80Listen()
		if !conf.Ssl.Only_EnableHttp2 {
			// 禁用h2
			ServerHttp.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}
		}
		ServerHttp.TLSConfig = &tls.Config{
			GetCertificate: conf.Ssl.GetNameToCert,
		}
		err = ServerHttp.ListenAndServeTLS("", "")
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
func Server80Listen() {
	if Server80Listening || !conf.Ssl.Old_EnableSsl || !conf.Ssl.Only_EnableListen80Redirect || conf.Http.Old_ServerHttpPortNumber != 443 {
		return
	}
	Server80Listening = true
	conf.Http.Old_Server80Port = conf.Http.New_Server80Port
	mylog.INFOln("http Start Server80Listen")
	Server80 = &http.Server{
		Addr:              conf.Http.Old_Server80Port,
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

// 正常停止所有服务
func ShutdownServerAll(inwg *sync.WaitGroup) {
	defer func() {
		if inwg != nil {
			inwg.Done()
		}
	}()
	mylog.INFOln("http ShutdownServerAll")
	var wg sync.WaitGroup
	wg.Add(1)
	go ShundownServer80(&wg)
	wg.Add(1)
	go ShundownServerHttp(&wg)
	wg.Wait()
}

// 正常停止80端口重定向http服务，包括tcp监听
func ShundownServer80(wg *sync.WaitGroup) {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()
	if !Server80Listening {
		return
	}
	mylog.INFOln("http ShundownServer80 , timeout 1s")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := Server80.Shutdown(ctx)
	if err != nil {
		mylog.ERRORln(err)
		CloseServer80()
		return
	}
	Server80 = nil
	Server80Listening = false
}

// 正常停止http服务，不包括tcp监听
func ShundownServerHttp(wg *sync.WaitGroup) {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()
	if !ServerHttpListening {
		return
	}
	mylog.INFOln("http ShundownServerHttp , timeout 10s")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := ServerHttp.Shutdown(ctx)
	if err != nil {
		mylog.ERRORln(err)
		CloseServerHttp()
		return
	}
	ServerHttp = nil
	ServerHttpListening = false
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
