package bcspcp

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

func Dial(SocketTempDir string) (*Context, error) {
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exPath := filepath.Dir(ex)

	dir := SocketTempDir
	if dir == "" {
		dir = "../" + DefaultSockDir
	}
	sockPath := filepath.Join(exPath, dir, DefaultSockName)

	c, err := net.Dial(NetworkName, sockPath)
	if err != nil {
		return nil, err
	}

	_, err = c.Write(ProtocolVersionHeader)
	if err != nil {
		c.Close()
		return nil, err
	}

	rb := make([]byte, 1)
	_, err = io.ReadFull(c, rb)
	if err != nil {
		c.Close()
		return nil, err
	}
	if rb[0] != 0 {
		c.Close()
		return nil, fmt.Errorf("bcspcp: protocol version %q does not supported", ProtocolVersionHeader)
	}

	return &Context{Conn: c}, nil
}
