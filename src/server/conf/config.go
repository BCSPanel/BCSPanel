package conf

import (
	"sync"
)

var readingLock sync.Mutex

func UpdateConfig() {
	Cmdport.UpdateConfig_cmdport()
	Http.UpdateConfig_http()
	Ssl.UpdateConfig_ssl()
}
