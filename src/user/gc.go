package user

import "time"

func gc() {
	for {
		time.Sleep(10 * time.Second)
		publicFuncLock.Lock()
		for name, c := range cache {
			if c.recentRequestTime.Add(cacheMaxAge).Before(time.Now()) {
				cache.del(name)
			}
		}
		publicFuncLock.Unlock()
	}
}
