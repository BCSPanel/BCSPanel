package config

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bddjr/BCSPanel/src/mylog"
	"github.com/bddjr/nametocert-go"
	"github.com/bytedance/sonic"
)

const userConfigHttpName = "http.json"

//go:embed default/http.json
var defaultHttpJson string

type http struct {
	Addr       string `json:"addr"`
	PathPrefix string `json:"path_prefix"`
	H2C        bool   `json:"h2c"`
	SSL        struct {
		Enable       bool `json:"enable"`
		Listen80Port bool `json:"listen_80_port"`
		Certs        []struct {
			Cert string `json:"cert"`
			Key  string `json:"key"`
		} `json:"certs"`
	} `json:"ssl"`

	AddrHost string
	AddrPort string
}

var OldHttp http
var NewHttp http
var HttpsCerts nametocert.Certs

func updateHttp() error {
	NewHttp = http{}

	// default
	err := sonic.UnmarshalString(defaultHttpJson, &NewHttp)
	if err != nil {
		return fmt.Errorf("unmarshal default: %v", err)
	}

	// user
	b, err := os.ReadFile(filepath.Join(userConfigDir, userConfigHttpName))
	if err != nil {
		if os.IsExist(err) {
			return err
		}
		create(userConfigHttpName, []byte(defaultHttpJson))
	} else {
		err = sonic.Unmarshal(b, &NewHttp)
	}

	if NewHttp.PathPrefix != "/" {
		NewHttp.PathPrefix = strings.ReplaceAll(NewHttp.PathPrefix, `\`, `/`)
		NewHttp.PathPrefix = path.Clean("/" + NewHttp.PathPrefix)
		if NewHttp.PathPrefix != "/" {
			NewHttp.PathPrefix += "/"
		}
	}

	if NewHttp.SSL.Enable {
		updateHttpsCerts()
	}

	if len(NewHttp.Addr) > 0 && NewHttp.Addr[0] == ':' {
		NewHttp.AddrPort = NewHttp.Addr[1:]
	} else {
		NewHttp.AddrHost, NewHttp.AddrPort, _ = net.SplitHostPort(NewHttp.Addr)
	}

	return err
}

func updateHttpsCerts() {
	HttpsCerts.Clear()

	for _, v := range NewHttp.SSL.Certs {
		certPath := filepath.Join(userConfigDir, v.Cert)
		keyPath := filepath.Join(userConfigDir, v.Key)
		// 证书与密钥
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			mylog.ERRORf("config UpdateHttpsCerts %q error: %v", certPath, err)
			continue
		}
		HttpsCerts.Add(&cert)
	}

	HttpsCerts.CompleteUpdate()
}

func UpdateHttp() error {
	mylog.INFO("config UpdateHttp")

	lock.Lock()
	defer lock.Unlock()

	err := updateHttp()
	if err != nil {
		mylog.ERROR("config UpdateHttp error: ", err)
	}
	return err
}

func SetHttpOld() {
	mylog.INFO("config SetHttpOld")
	OldHttp = NewHttp
}
