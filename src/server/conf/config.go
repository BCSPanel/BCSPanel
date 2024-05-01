package conf

import (
	"sync"

	"github.com/bddjr/BCSPanel/src/server/mylog"
	"github.com/spf13/viper"
)

var readingLock sync.Mutex

type ConfigConfigType struct {
	ColorScheme string
}

var ConfigConfig = &ConfigConfigType{}

func UpdateConfig() {
	Cmdport.UpdateConfig_cmdport()
	Http.UpdateConfig_http()
	Robots.UpdateConfig_robots()
	Ssl.UpdateConfig_ssl()
	ConfigConfig.UpdateConfig_config()
}

func ToUint16(n int) uint16 {
	return uint16(max(0, min(n, 65535)))
}

func (c *ConfigConfigType) UpdateConfig_config() {
	readingLock.Lock()
	defer readingLock.Unlock()
	mylog.INFOln("updateConfig_config")
	// 读config
	viper.SetConfigFile("./conf/config.yml")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		mylog.ERRORln(err)
		c.ColorScheme = ""
		return
	}

	c.ColorScheme, _ = viper.Get("color_scheme").(string)
}
