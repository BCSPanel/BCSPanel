package cmdserver

import (
	"bytes"
	"crypto/hmac"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"strings"
	"time"

	"github.com/bddjr/BCSPanel/src/cmdserver/sharecmdlistener"
	"github.com/bddjr/BCSPanel/src/conf"
	"github.com/bddjr/BCSPanel/src/myhmac"
	"github.com/bddjr/BCSPanel/src/mylog"
	"github.com/bddjr/BCSPanel/src/myrand"
	"github.com/bddjr/BCSPanel/src/myversion"
	"github.com/bddjr/BCSPanel/src/reload"
	"github.com/bddjr/BCSPanel/src/shutdown"
	"github.com/bddjr/BCSPanel/src/user"
)

type CommandType struct {
}

var keyForServer [64]byte
var keyForClient [64]byte

var listener net.Listener = nil

const ListenIP = "127.209.156.239"
const cacheFileMode = fs.FileMode(0600)

func Start() {
	go start()
}

// 会阻塞进程，不想阻塞请用go
func start() {
	mylog.INFOf("cmdserver Start port %d\n", conf.Cmdport.NewPort)
	conf.Cmdport.OldPort = conf.Cmdport.NewPort

	// 生成Key
	myrand.Read(keyForClient[:])
	myrand.Read(keyForServer[:])
	// 写入文件
	os.Mkdir("cache", cacheFileMode)
	err := os.WriteFile("cache/cmd-client-key", keyForClient[:], cacheFileMode)
	if err != nil {
		mylog.ERRORln("cmdport write file error:", err)
		return
	}
	err = os.WriteFile("cache/cmd-server-key", keyForServer[:], cacheFileMode)
	if err != nil {
		mylog.ERRORln("cmdport write file error:", err)
		return
	}
	err = os.WriteFile("cache/cmd-port", []byte(fmt.Sprint(conf.Cmdport.OldPort)), cacheFileMode)
	if err != nil {
		mylog.ERRORln("cmdport write file error:", err)
		return
	}

	listener, err = net.Listen("tcp4", fmt.Sprint(ListenIP, ":", conf.Cmdport.OldPort))
	if err != nil {
		if err != net.ErrClosed {
			mylog.ERRORln("cmdport listen error:", err)
		}
		return
	}
	defer func() {
		if sharecmdlistener.Reloading {
			go start()
		} else {
			listener.Close()
		}
	}()

	sharecmdlistener.Listener = &listener
	for {
		c, err := listener.Accept()
		if err != nil {
			// mylog.ERRORln("cmdport accept error:", err)
			break
		}
		go handleConn(c)
	}
}

func setReadTimeout10s(c net.Conn) {
	c.SetReadDeadline(time.Now().Add(10 * time.Second))
}

func setReadTimeout3s(c net.Conn) {
	c.SetReadDeadline(time.Now().Add(3 * time.Second))
}

func setReadTimeout0(c net.Conn) {
	c.SetReadDeadline(time.Time{})
}

func readJson(c net.Conn) (jsonStr string, err error) {
	// 写法参考了 io.ReadAll
	b := make([]byte, 0, 512)
	for {
		var n int
		setReadTimeout3s(c)
		n, err = c.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			return
		}

		// 当结尾是 } 与 0x03 ，则认为传输完成
		if bytes.HasSuffix(b, []byte{'}', 0x03}) {
			jsonStr = string(b)
			return
		}

		// 分配更多空间
		if len(b) == cap(b) {
			b = append(b, 0)[:len(b)]
		}
	}
}

func handleConn(c net.Conn) {
	// 函数退出时自动关闭连接
	defer c.Close()

	// 不是127网段就滚
	if !strings.HasPrefix(c.RemoteAddr().String(), "127.") {
		return
	}

	{
		// 读取客户端发送的信息头必须是"BCSPCP-1.0>"，后面带有盐
		const head = "BCSPCP-1.0>"
		const headLen = len(head)
		const saltLen = 16
		var rb [headLen + saltLen]byte
		setReadTimeout3s(c)
		n, err := c.Read(rb[:])
		if err != nil || n < headLen+saltLen || !bytes.Equal(rb[:headLen], []byte(head)) {
			mylog.ERRORln("cmdport read head from client failed, err:", err)
			return
		}
		// 加密发送服务器认证密码给客户端
		_, err = c.Write(myhmac.HmacSha3_512(keyForServer[:], rb[headLen:]))
		if err != nil {
			mylog.ERRORln("cmdport sent failed, err:", err)
			return
		}
	}
	// 验证客户端的密码，读指令编号
	const passwdLen = 64
	var rb [passwdLen + 1]byte
	setReadTimeout3s(c)
	n, err := c.Read(rb[:])
	if err != nil || n < passwdLen || !hmac.Equal(rb[:passwdLen], keyForClient[:]) {
		mylog.ERRORln("cmdport read from client failed, cmdkey err:", err)
		return
	}
	setReadTimeout0(c)
	cmdNum := rb[passwdLen]
	switch cmdNum {
	case 1:
		// 关闭程序
		io.WriteString(c, "BCSPanel Shutting Down\n")
		shutdown.Shutdown(0)
	case 2:
		// 重载
		io.WriteString(c, "BCSPanel Reloading\n")
		reload.Reload()
	case 3:
		// 版本
		io.WriteString(c, fmt.Sprint("BCSPanel Version ", myversion.Version, "\n"))
	case 4:
		// 刷新并获取注册验证码
		io.WriteString(c, fmt.Sprint(
			"BCSPanel Register Verify Code:\n\n  ", user.RegisterVerifyCode.Fill().Code, "\n\n",
			"Expiration Time: ", user.RegisterVerifyCode.ExpirationTime.Format(user.TimeFormat), "\n",
		))
	case 5:
		// 更新
	case 6:
		// 从用户名获取用户id
		nameBytes := make([]byte, 32)
		n, err := c.Read(nameBytes)
		if err != nil {
			io.WriteString(c, fmt.Sprint("Error: ", err, "\n"))
			return
		}
		nameBytes = nameBytes[:n]
		id, err := user.UserNameToId(string(nameBytes))
		if err != nil {
			io.WriteString(c, fmt.Sprint("Error: ", err, "\n"))
			return
		}
		io.WriteString(c, fmt.Sprint(id, "\n"))
	case 7:
		// 更改用户名
	case 8:
		// 更改密码
	default:
		io.WriteString(c, "Error: Unkown Command\n")
	}
}
