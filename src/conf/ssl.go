package conf

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"path"
	"strings"

	"github.com/bddjr/BCSPanel/src/mylog"
	"github.com/spf13/viper"
)

type ConfigSslCertsType map[string]*tls.Certificate

type ConfigSslType struct {
	NameToCerts ConfigSslCertsType

	Old_EnableSsl bool
	New_EnableSsl bool

	Only_EnableListen80Redirect bool
	Only_EnableHttp2            bool
}

var Ssl = &ConfigSslType{}

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

	// 新证书表
	newCerts := ConfigSslCertsType{}
	defer func() {
		c.NameToCerts = newCerts
	}()

	// 读默认证书
	cert, err := tls.LoadX509KeyPair(
		"./src/server/httpserver/cert/localhost.crt",
		"./src/server/httpserver/cert/localhost.key",
	)
	if err != nil {
		mylog.ERRORln("Can not read default cert")
	} else {
		newCerts["*"] = &cert
	}

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
		// 解析证书
		Certificate, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			mylog.ERRORf("Can not parse cert \"%s\" , %v\n", certName, err)
			continue
		}
		// 依据证书里的可选名称，自动匹配
		names := Certificate.DNSNames
		for k := 0; k < len(names); k++ {
			name := names[k]
			newCerts[name] = &cert
		}
		// 如果证书有可选IP，那么当servername为空的时候匹配
		if len(Certificate.IPAddresses) > 0 {
			newCerts[""] = &cert
		}
	}
}

func (c *ConfigSslType) GetNameToCert(ClientHelloInfo *tls.ClientHelloInfo) (cert *tls.Certificate, err error) {
	// www.example.com
	cert, ok := c.NameToCerts[ClientHelloInfo.ServerName]
	if ok {
		return
	}
	// *.example.com
	if i := strings.IndexByte(ClientHelloInfo.ServerName, '.'); i != -1 {
		cert, ok = c.NameToCerts["*"+ClientHelloInfo.ServerName[i:]]
		if ok {
			return
		}
	}
	// default
	cert, ok = c.NameToCerts["*"]
	if !ok {
		err = fmt.Errorf(`can not find cert for "%s"`, ClientHelloInfo.ServerName)
	}
	return
}
