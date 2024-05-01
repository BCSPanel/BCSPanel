package conf

import (
	"github.com/bddjr/BCSPanel/src/server/mylog"
	"github.com/spf13/viper"
)

type ConfigRobotsType struct {
	EnableXRobotsTag bool
	EnableRobotsTxt  bool
}

var Robots = &ConfigRobotsType{}

func (c *ConfigRobotsType) UpdateConfig_robots() {
	readingLock.Lock()
	defer readingLock.Unlock()
	mylog.INFOln("updateConfig_robots")
	// 读robots
	viper.SetConfigFile("./conf/robots.yml")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		mylog.ERRORln(err)
		c.EnableXRobotsTag = true
		c.EnableRobotsTxt = false
		return
	}

	var ok bool
	// X-Robots-Tag 响应头，默认开
	c.EnableXRobotsTag, ok = viper.Get("enable_x_robots_tag").(bool)
	if !ok {
		c.EnableXRobotsTag = true
	}

	// robots.txt 默认关
	c.EnableRobotsTxt, _ = viper.Get("enable_robots_txt").(bool)
}
