package bcspcp

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

func setReadTimeout10s(c net.Conn) {
	c.SetReadDeadline(time.Now().Add(10 * time.Second))
}

func setReadTimeout0(c net.Conn) {
	c.SetReadDeadline(time.Time{})
}

type Message map[string]any

type ErrorLogf func(format string, v ...any)

type Server struct {
	Handler   func(ctx *Context) error
	ErrorLogf ErrorLogf

	// Default: ".bcspcp"
	SocketTempDir string

	closeListener func() error
}

func (srv *Server) logf(format string, v ...any) {
	if srv.ErrorLogf != nil {
		srv.ErrorLogf(format, v...)
	} else {
		log.Printf(format, v...)
	}
}

func (srv *Server) getDirName() string {
	if srv.SocketTempDir != "" {
		return srv.SocketTempDir
	}
	return DefaultSockDir
}

func (srv *Server) Close() {
	if srv.closeListener != nil {
		srv.closeListener()
		srv.closeListener = nil
		os.Remove(srv.getDirName())
	}
}

func (srv *Server) ListenAndServe() error {
	if srv.closeListener != nil {
		return errors.New("bcspcp: listening")
	}

	dir := srv.getDirName()
	sockPath := filepath.Join(dir, DefaultSockName)

	err := os.MkdirAll(dir, SockFilePerm)
	if err != nil {
		if err == os.ErrExist {
			err = os.Chmod(dir, SockFilePerm)
		}
		if err != nil {
			return err
		}
	}
	os.Remove(sockPath)

	listener, err := net.Listen(NetworkName, sockPath)
	if err != nil {
		return err
	}
	srv.closeListener = listener.Close
	defer func() { srv.closeListener = nil }()
	defer listener.Close()

	for {
		c, err := listener.Accept()
		if err != nil {
			return err
		}
		go srv.serve(c)
	}
}

func (srv *Server) serve(c net.Conn) {
	defer c.Close()
	defer func() {
		if err := recover(); err != nil {
			srv.logf("bcspcp: panic error: %v", err)
		}
	}()

	b := make([]byte, len(ProtocolVersionHeader))
	setReadTimeout10s(c)
	_, err := io.ReadFull(c, b)
	if err != nil {
		srv.logf("bcspcp: read error: %v", err)
		return
	}

	setReadTimeout0(c)
	if !bytes.Equal(b, ProtocolVersionHeader) {
		c.Write([]byte{1})
		srv.logf("bcspcp: read header %q does not supported", b)
		return
	}
	c.Write([]byte{0})

	err = srv.Handler(&Context{Conn: c})
	if err != nil {
		srv.logf("bcspcp: handler error: %v", err)
	}
}
