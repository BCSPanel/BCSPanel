package mysession

import "time"

func gc() {
	for {
		time.Sleep(10 * time.Minute)
		lock.Lock()
		// 遍历
		for _, s := range sessions {
			if s.LastUsageTime.Add(MaxAge).Before(time.Now()) {
				removeSession(s)
			}
		}
		lock.Unlock()
	}
}
