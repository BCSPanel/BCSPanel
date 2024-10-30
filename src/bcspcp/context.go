package bcspcp

import (
	"net"

	"github.com/bytedance/sonic"
)

type Context struct {
	Conn net.Conn
}

func (ctx *Context) ReadMsg() (Message, error) {
	msg := make(Message)
	err := sonic.ConfigDefault.NewDecoder(ctx.Conn).Decode(&msg)
	return msg, err
}

func (ctx *Context) WriteMsg(msg Message) error {
	return sonic.ConfigDefault.NewEncoder(ctx.Conn).Encode(&msg)
}

func (ctx *Context) Close() error {
	return ctx.Conn.Close()
}
