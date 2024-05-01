package conf

import (
	"github.com/bddjr/BCSPanel/src/server/isservermode"
	"github.com/bddjr/BCSPanel/src/server/mylog"
	"github.com/spf13/viper"
)

type ConfigCmdportType struct {
	OldPort uint16
	NewPort uint16
}

var Cmdport = &ConfigCmdportType{}

func (c *ConfigCmdportType) UpdateConfig_cmdport() {
	readingLock.Lock()
	defer readingLock.Unlock()
	if isservermode.IsServerMode {
		mylog.INFOln("updateConfig_cmdport")
	}
	// 读robots
	viper.SetConfigFile("./conf/cmdport.yml")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		mylog.ERRORln(err)
		c.OldPort = 39625
		return
	}

	port, ok := viper.Get("cmd_port").(int)
	if !ok {
		c.NewPort = 39625
	} else {
		c.NewPort = ToUint16(port)
	}
}
