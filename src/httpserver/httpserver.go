package httpserver

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bddjr/BCSPanel/src/conf"
	"github.com/bddjr/BCSPanel/src/httprouter"
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
		sb.WriteString(fmt.Sprint(":", conf.Http.New_ServerHttpPortNumber))
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
		if !conf.Ssl.New_EnableSsl ||
			!conf.Ssl.Only_EnableListen80Redirect ||
			conf.Http.New_ServerHttpPortNumber != 443 {
			CloseServer80()
		} else if conf.Http.Old_Server80Addr != conf.Http.New_Server80Addr {
			// 监听端口变了
			CloseServer80()
			go Server80Listen()
		}
	} else if conf.Ssl.Only_EnableListen80Redirect {
		// 未监听，如果设为true就开始监听
		go Server80Listen()
	}
	// 当监听端口变了
	if conf.Http.Old_ServerHttpAddr != conf.Http.New_ServerHttpAddr ||
		// 改变了ssl开启状态
		conf.Ssl.Old_EnableSsl != conf.Ssl.New_EnableSsl ||
		// 更改了保持连接时长
		conf.Http.Old_KeepAliveSecond != conf.Http.New_KeepAliveSecond ||
		// 更改了gzip压缩等级
		conf.Http.Old_GzipLevel != conf.Http.New_GzipLevel ||
		// 改变了H2C开启状态
		(!conf.Ssl.New_EnableSsl && conf.Http.Old_EnableH2c != conf.Http.New_EnableH2c) ||
		// 更改了路径开头
		conf.Http.Old_PathPrefix != conf.Http.New_PathPrefix ||
		// 改变了Basic登录的开启状态
		conf.Http.Old_EnableBasicLogin != conf.Http.New_EnableBasicLogin ||
		// 改变了未知名称拒绝握手的开启状态
		conf.Ssl.Old_EnableRejectHandshakeIfUnrecognizedName != conf.Ssl.New_EnableRejectHandshakeIfUnrecognizedName {
		// 那么
		// 重启 ServerHttp
		mylog.INFOln("http Reload ServerHttp")
		CloseServerHttp()
		go ServerHttpListen()
		return
	}
	if conf.Ssl.New_EnableSsl {
		// 如果已开启https，检查是否启用http2
		if conf.Ssl.Only_EnableHttp2 {
			// 启用
			ServerHttp.TLSNextProto = nil
		} else {
			// 禁用
			ServerHttp.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}
		}
	}
}

// 启动服务，包括tcp监听。
// 会自动异步启动80端口重定向服务 Server80Listen 。
// 该函数会阻塞运行。
func ServerHttpListen() {
	if ServerHttpListening {
		return
	}
	ServerHttpListening = true
	conf.Http.Old_ServerHttpAddr = conf.Http.New_ServerHttpAddr
	conf.Http.Old_ServerHttpPortNumber = conf.Http.New_ServerHttpPortNumber
	conf.Ssl.Old_EnableSsl = conf.Ssl.New_EnableSsl
	conf.Http.Old_KeepAliveSecond = conf.Http.New_KeepAliveSecond
	conf.Ssl.Old_EnableRejectHandshakeIfUnrecognizedName = conf.Ssl.New_EnableRejectHandshakeIfUnrecognizedName

	mylog.INFOf("http ServerHttpListen port %s , ssl %v\n", conf.Http.Old_ServerHttpAddr, conf.Ssl.Old_EnableSsl)
	ServerHttp = hlfhr.New(&http.Server{
		Addr:              conf.Http.Old_ServerHttpAddr,
		Handler:           httprouter.GetHandler(),
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
		conf.Ssl.CertsProc.RejectHandshakeIfUnrecognizedName = conf.Ssl.Old_EnableRejectHandshakeIfUnrecognizedName
		ServerHttp.TLSConfig = &tls.Config{
			GetCertificate: conf.Ssl.CertsProc.GetCertificate,
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
	conf.Http.Old_Server80Addr = conf.Http.New_Server80Addr
	mylog.INFOln("http Start Server80Listen")
	Server80 = &http.Server{
		Addr:              conf.Http.Old_Server80Addr,
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
