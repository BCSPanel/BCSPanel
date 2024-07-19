package conf

import (
	"crypto/tls"
	"path"

	"github.com/bddjr/BCSPanel/src/mylog"
	"github.com/bddjr/nametocert-go"
	"github.com/spf13/viper"
)

type ConfigSslCertsType map[string]*tls.Certificate

type ConfigSslType struct {
	CertsProc nametocert.Processor

	Old_EnableSsl bool
	New_EnableSsl bool

	Only_EnableListen80Redirect bool
	Only_EnableHttp2            bool

	Old_EnableRejectHandshakeIfUnrecognizedName bool
	New_EnableRejectHandshakeIfUnrecognizedName bool
}

var Ssl ConfigSslType

func (c *ConfigSslType) UpdateConfig_ssl() {
	readingLock.Lock()
	defer readingLock.Unlock()
	mylog.INFOln("updateConfig_ssl")
	// 读配置文件
	viper.SetConfigFile("./conf/ssl.yml")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		mylog.ERRORln(err)
		c.New_EnableSsl = false
		return
	}

	// 是否开启SSL，默认关
	c.New_EnableSsl, _ = viper.Get("enable").(bool)
	if !c.New_EnableSsl {
		// 关，这里不需要更改Certs
		return
	}

	// 监听80端口用于重定向，默认开
	var ok bool
	c.Only_EnableListen80Redirect, ok = viper.Get("enable_listen_80_redirect").(bool)
	if !ok {
		c.Only_EnableListen80Redirect = true
	}

	// 开启http2，仅开启HTTPS时有效，默认开
	c.Only_EnableHttp2, ok = viper.Get("enable_http2").(bool)
	if !ok {
		c.Only_EnableHttp2 = true
	}

	// 如果找不到与名称匹配的证书，拒绝握手。
	c.New_EnableRejectHandshakeIfUnrecognizedName, _ = viper.Get("enable_reject_handshake_if_unrecognized_name").(bool)

	// 新证书表
	certs := nametocert.Certs{}
	defer c.CertsProc.SetCerts(certs)

	// 证书合集
	ymlCerts, ok := viper.Get("certs").([]interface{})
	if !ok {
		return
	}
	for _, v := range ymlCerts {
		item := v.(map[string]interface{})
		certName := item["cert"].(string)
		certPath := path.Join("./conf/cert", certName)
		keyName := item["key"].(string)
		keyPath := path.Join("./conf/cert", keyName)
		// 证书与密钥
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			mylog.ERRORf("Can not read cert \"%s\" , %v\n", certName, err)
			continue
		}
		certs.Add(&cert)
	}
}
