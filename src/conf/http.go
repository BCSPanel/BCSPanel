package conf

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bddjr/BCSPanel/src/mylog"
	"github.com/spf13/viper"
)

type ConfigHttpType struct {
	// ":24124"
	Old_ServerHttpAddr string
	New_ServerHttpAddr string

	// 24124
	Old_ServerHttpPortNumber uint16
	New_ServerHttpPortNumber uint16

	// ":80"
	Old_Server80Addr string
	New_Server80Addr string

	// 5
	Old_GzipLevel int
	New_GzipLevel int

	// 1024
	Old_GzipMinContentLength int64
	New_GzipMinContentLength int64

	// "/"
	Old_PathPrefix string
	New_PathPrefix string

	Only_AddHeaders []map[string]string

	Only_EnableGinLog bool

	Old_EnableBasicLogin bool
	New_EnableBasicLogin bool

	Old_KeepAliveSecond int
	New_KeepAliveSecond int

	Only_EnableXRealIp  bool
	Only_XForwarderAuth string

	Old_EnableH2c bool
	New_EnableH2c bool
}

var Http = &ConfigHttpType{}

func (c *ConfigHttpType) UpdateConfig_http() {
	readingLock.Lock()
	defer readingLock.Unlock()
	mylog.INFOln("updateConfig_http")
	// 读config
	viper.SetConfigFile("./conf/http.yml")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		mylog.ERRORln(err)
		c.New_ServerHttpAddr = ":24124"
		c.New_ServerHttpPortNumber = 24124
		c.New_Server80Addr = ":80"
		c.New_GzipLevel = 5
		c.New_GzipMinContentLength = 1024
		c.New_PathPrefix = "/"
		c.Only_AddHeaders = nil
		c.New_EnableBasicLogin = false
		c.Only_EnableGinLog = true
		c.New_KeepAliveSecond = 180
		c.New_EnableH2c = false
		return
	}

	// http服务tcp监听端口，默认24124
	c.New_Server80Addr = ":80"
	if addr := viper.GetUint16("listen"); addr != 0 {
		// int类型
		c.New_ServerHttpPortNumber = addr
		c.New_ServerHttpAddr = fmt.Sprint(":", addr)
	} else {
		// 可能是string类型
		c.New_ServerHttpAddr = viper.GetString("listen")
		regexpPort := regexp.MustCompile(`:\d+$`)
		if find := regexpPort.FindString(c.New_ServerHttpAddr); find != "" {
			// ":24124" "0.0.0.0:24124"
			n, _ := strconv.Atoi(find[1:])
			c.New_ServerHttpPortNumber = uint16(n)
			// "0.0.0.0:80"
			c.New_Server80Addr = regexpPort.ReplaceAllString(c.New_ServerHttpAddr, ":80")
		} else if n, err := strconv.Atoi(c.New_ServerHttpAddr); err == nil {
			// "24124"
			c.New_ServerHttpPortNumber = uint16(n)
		} else {
			// "0.0.0.0" ""
			c.New_ServerHttpPortNumber = 24124
			c.New_Server80Addr = c.New_ServerHttpAddr + ":80"
			c.New_ServerHttpAddr += ":24124"
		}
	}

	// http服务Gzip压缩等级，默认5
	var ok bool
	c.New_GzipLevel, ok = viper.Get("gzip_level").(int)
	if !ok {
		c.New_GzipLevel = 5
	}
	if c.New_GzipLevel != 0 {
		// Gzip最小需要到达多少字节才会压缩，默认1024
		c.New_GzipMinContentLength, ok = viper.Get("gzip_min_content_length").(int64)
		if !ok {
			c.New_GzipMinContentLength = 1024
		}
	}

	// http服务路由前缀
	c.New_PathPrefix, _ = viper.Get("path_prefix").(string)
	if !strings.HasPrefix(c.New_PathPrefix, "/") {
		c.New_PathPrefix = "/" + c.New_PathPrefix
	}
	if !strings.HasSuffix(c.New_PathPrefix, "/") {
		c.New_PathPrefix += "/"
	}
	if b, _ := regexp.MatchString(`^/[0-9a-zA-Z]+/$`, c.New_PathPrefix); !b {
		c.New_PathPrefix = "/"
	}

	// 添加响应头
	if addHeaders, ok := viper.Get("add_headers").([]interface{}); ok {
		newHeaders := make([]map[string]string, len(addHeaders))
		for k, v := range addHeaders {
			item := v.(map[string]interface{})
			newHeaders[k] = make(map[string]string)
			h := newHeaders[k]
			for k, v := range item {
				h[k] = fmt.Sprint(v)
			}
		}
		c.Only_AddHeaders = newHeaders
	} else {
		c.Only_AddHeaders = nil
	}

	// 使用Basic登录
	c.New_EnableBasicLogin, _ = viper.Get("enable_basic_login").(bool)

	// http服务打印Gin框架的日志
	c.Only_EnableGinLog, ok = viper.Get("enable_gin_log").(bool)
	if !ok {
		c.Only_EnableGinLog = true
	}

	// http服务keepAlive时长（秒），默认180
	c.New_KeepAliveSecond, ok = viper.Get("keep_alive_second").(int)
	if !ok {
		c.New_KeepAliveSecond = 180
	}

	// 启用识别客户端IP优先采用请求头 X-Real-Ip
	// 默认false
	c.Only_EnableXRealIp, _ = viper.Get("enable_x_real_ip").(bool)

	// 反向代理身份验证请求头 X-Forwarder-Auth-Bcspanel 的内容
	// 默认空
	c.Only_XForwarderAuth, _ = viper.Get("x_forwarder_auth").(string)

	// HTTP/2 H2C 默认关
	c.New_EnableH2c, _ = viper.Get("enable_h2c").(bool)
}
