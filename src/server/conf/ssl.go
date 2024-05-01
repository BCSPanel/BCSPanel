package conf

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"path"
	"regexp"

	"github.com/bddjr/BCSPanel/src/server/mylog"
	"github.com/spf13/viper"
)

type ConfigSslCertsType map[string]*tls.Certificate

type ConfigSslType struct {
	Certs ConfigSslCertsType

	Old_EnableSsl bool
	New_EnableSsl bool

	Only_EnableUnknownServernameSend421 bool
	Only_EnableListen80Redirect         bool
	Only_EnableHttp2                    bool

	Only_HSTS string
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

	// 新证书表
	newCerts := make(ConfigSslCertsType, 0)

	// 读默认证书
	cert, err := tls.LoadX509KeyPair(
		"./src/server/httpserver/cert/localhost.crt",
		"./src/server/httpserver/cert/localhost.key",
	)
	if err != nil {
		fmt.Println("Can not read default cert")
	} else {
		newCerts["*"] = &cert
	}

	// 监听80端口用于重定向，默认开
	var ok bool
	c.Only_EnableListen80Redirect, ok = viper.Get("enable_listen_80_redirect").(bool)
	if !ok {
		c.Only_EnableListen80Redirect = true
	}

	// HSTS，默认不添加
	c.Only_HSTS, _ = viper.Get("hsts").(string)

	// 未知servername返回421状态码，默认关
	c.Only_EnableUnknownServernameSend421, _ = viper.Get("enable_unknown_servername_send_421").(bool)

	// 开启http2，仅开启HTTPS时有效，默认开
	c.Only_EnableHttp2, ok = viper.Get("enable_http2").(bool)
	if !ok {
		c.Only_EnableHttp2 = true
	}

	// 证书合集
	ymlCerts, ok := viper.Get("certs").([]interface{})
	if !ok {
		c.Certs = newCerts
		return
	}
	for i := 0; i < len(ymlCerts); i++ {
		j := ymlCerts[i].(map[string]interface{})
		certPath := path.Join("./conf/cert", j["cert"].(string))
		keyPath := path.Join("./conf/cert", j["key"].(string))
		// 证书与密钥
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			mylog.ERRORf("Can not read cert \"%s\" , %v\n", j["cert"].(string), err)
			continue
		}
		// 解析证书
		Certificate, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			mylog.ERRORf("Can not parse cert \"%s\" , %v\n", j["cert"].(string), err)
			continue
		}
		// 依据证书里的可选名称，自动匹配
		names := Certificate.DNSNames
		for k := 0; k < len(names); k++ {
			name := names[k]
			newCerts[name] = &cert
		}
		// 如果证书有可选IP，那么当servername为空的时候匹配
		ips := Certificate.IPAddresses
		if len(ips) > 0 {
			newCerts[""] = &cert
			// 这块一般没有实际作用，主要是给 isUnknownServername 准备的
			for k := 0; k < len(ips); k++ {
				newCerts[ips[k].String()] = &cert
			}
		}
	}
	c.Certs = newCerts
}

var compiledRegExp_tlsServerName = regexp.MustCompile(`^[^\.]+`)

func (c *ConfigSslType) getNameToCert(name string, UnknownGetDefault bool) (*tls.Certificate, bool, string) {
	// www.example.com
	cert, ok := c.Certs[name]
	if ok {
		return cert, true, name
	}
	// *.example.com
	_name := compiledRegExp_tlsServerName.ReplaceAllString(name, `*`)
	cert, ok = c.Certs[_name]
	if ok {
		return cert, true, _name
	}
	// default
	if UnknownGetDefault {
		cert, ok = c.Certs["*"]
		if ok {
			return cert, true, "*"
		}
	}
	return nil, false, ""
}

func Ssl_GetCertificate(ClientHelloInfo *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert, ok, _ := Ssl.getNameToCert(ClientHelloInfo.ServerName, true)
	if ok {
		return cert, nil
	}
	return nil, errors.New(`can not find cert for "` + ClientHelloInfo.ServerName + `"`)
}

func Ssl_IsUnknownServername(Hostname string) bool {
	if Hostname == "" {
		return true
	}
	_, ok, name := Ssl.getNameToCert(Hostname, false)
	if name == "*" {
		return true
	}
	return !ok
}

func Ssl_NeedToReturn421ForUnknownServername(Hostname string) bool {
	if !Ssl.Only_EnableUnknownServernameSend421 {
		return false
	}
	return Ssl_IsUnknownServername(Hostname)
}
