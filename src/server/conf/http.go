package conf

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bddjr/BCSPanel/src/server/mylog"
	"github.com/spf13/viper"
)

type ConfigHttpType struct {
	// ":24124"
	Old_ServerHttpPort string
	New_ServerHttpPort string

	// 24124
	Old_ServerHttpPortNumber uint16
	New_ServerHttpPortNumber uint16

	// ":80"
	Old_Server80Port string
	New_Server80Port string

	// 5
	Old_GzipLevel int16
	New_GzipLevel int16

	// 1024
	Old_GzipMinContentLength int64
	New_GzipMinContentLength int64

	// "/"
	Old_PathPrefix string
	New_PathPrefix string

	Only_EnableGinLog bool

	Old_KeepAliveSecond uint
	New_KeepAliveSecond uint

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
		c.New_ServerHttpPort = ":24124"
		c.New_ServerHttpPortNumber = 24124
		c.New_Server80Port = ":80"
		c.New_GzipLevel = 5
		c.New_GzipMinContentLength = 1024
		c.New_PathPrefix = "/"
		c.Only_EnableGinLog = true
		c.New_KeepAliveSecond = 10
		c.New_EnableH2c = false
		return
	}

	// http服务tcp监听端口，默认24124
	c.New_Server80Port = ":80"
	if _ServerHttpPort, ok := viper.Get("listen").(int); ok {
		// int类型
		c.New_ServerHttpPortNumber = ToUint16(_ServerHttpPort)
		c.New_ServerHttpPort = fmt.Sprint(":", c.New_ServerHttpPortNumber)
	} else {
		// 可能是string类型
		c.New_ServerHttpPort, _ = viper.Get("listen").(string)
		regexpPort := regexp.MustCompile(`:\d+$`)
		if find := regexpPort.FindString(c.New_ServerHttpPort); find != "" {

			// ":24124" "0.0.0.0:24124"
			n, _ := strconv.Atoi(find[1:])
			c.New_ServerHttpPortNumber = ToUint16(n)
			// ":99999" -> ":65535"
			c.New_ServerHttpPort = regexpPort.ReplaceAllString(c.New_ServerHttpPort, fmt.Sprint(":", c.New_ServerHttpPortNumber))
			// "0.0.0.0:80"
			c.New_Server80Port = regexpPort.ReplaceAllString(c.New_ServerHttpPort, ":80")

		} else if matched, _ := regexp.MatchString(`^\d+$`, c.New_ServerHttpPort); matched {

			// "24124"
			n, _ := strconv.Atoi(c.New_ServerHttpPort)
			c.New_ServerHttpPortNumber = ToUint16(n)
			// ":24124"
			c.New_ServerHttpPort = fmt.Sprint(":", c.New_ServerHttpPort)

		} else {

			// "0.0.0.0" ""
			c.New_ServerHttpPortNumber = 24124
			c.New_Server80Port = c.New_ServerHttpPort + ":80"
			c.New_ServerHttpPort += ":24124"

		}
	}

	// http服务Gzip压缩等级，默认5
	var ok bool
	c.New_GzipLevel, ok = viper.Get("gzip_level").(int16)
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
	if !strings.HasSuffix(c.New_PathPrefix, "/") {
		c.New_PathPrefix += "/"
	}
	if !strings.HasPrefix(c.New_PathPrefix, "/") {
		c.New_PathPrefix = "/" + c.New_PathPrefix
	}
	if b, _ := regexp.MatchString(`^/[0-9a-zA-Z]+/$`, c.New_PathPrefix); !b {
		c.New_PathPrefix = "/"
	}

	// http服务打印Gin框架的日志
	c.Only_EnableGinLog, ok = viper.Get("enable_gin_log").(bool)
	if !ok {
		c.Only_EnableGinLog = true
	}

	// http服务keepAlive时长（秒），默认10
	c.New_KeepAliveSecond, ok = viper.Get("keep_alive_second").(uint)
	if !ok {
		c.New_KeepAliveSecond = 10
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
