package cmdclient

import (
	"crypto/hmac"
	"fmt"
	"net"
	"os"

	"github.com/bddjr/BCSPanel/src/cmdserver"
	"github.com/bddjr/BCSPanel/src/myhmac"
	"github.com/bddjr/BCSPanel/src/myrand"
	"github.com/bddjr/BCSPanel/src/myversion"
)

func Run() {
	if len(os.Args) == 1 || os.Args[1] == "help" {
		fmt.Println("BCSPanel Command Help")
		return
	}

	var cmdId byte
	switch os.Args[1] {
	case "stop":
		cmdId = 1
	case "reload":
		cmdId = 2
	case "version":
		// cmdId = 3
		fmt.Println("BCSPanel Version", myversion.Version)
		return
	case "verify":
		cmdId = 4
	case "usernametoid":
		cmdId = 6
	default:
		fmt.Println("Error: Unkown Command")
		return
	}

	var port string
	{
		fb, err := os.ReadFile("cache/cmd-port")
		if err != nil {
			fmt.Println("Error: BCSPanel Not Running!")
			fmt.Println(err)
			return
		}
		port = fmt.Sprint(cmdserver.ListenIP, ":", string(fb))
	}
	c, err := net.Dial("tcp4", port)
	if err != nil {
		fmt.Println("Error: BCSPanel Not Running!")
		fmt.Println(err)
		return
	}
	defer c.Close()

	// 协议头+盐
	head := []byte("BCSPCP-1.0>")
	salt := myrand.RandBytes(16)
	_, err = c.Write(append(head, salt...))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// 验证服务器的认证码
	{
		fb, err := os.ReadFile("cache/cmd-server-key")
		if err != nil {
			fmt.Println(err)
			return
		}
		nb := make([]byte, len(fb))
		_, err = c.Read(nb)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if !hmac.Equal(nb, myhmac.HmacSha3_512(fb, salt)) {
			fmt.Println("Error: Server not authenticated")
			return
		}
	}
	// 发送客户端验证码与命令
	{
		fb, err := os.ReadFile("cache/cmd-client-key")
		if err != nil {
			fmt.Println(err)
			return
		}
		b := append(fb, cmdId)
		if cmdId == 6 {
			b = append(b, []byte(os.Args[2])...)
		}
		c.Write(b)
	}
	// 接收服务器返回内容，然后输出
	for {
		nb := make([]byte, 4096)
		n, err := c.Read(nb)
		if err != nil {
			// fmt.Println("Error:", err)
			return
		}
		nb = nb[:n]
		os.Stdout.Write(nb)
	}

	// 等待服务器关闭连接
	// c.Read(make([]byte, 1))
}
