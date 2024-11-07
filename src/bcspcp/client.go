package bcspcp

import (
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

	return &Context{Conn: c}, nil
}
