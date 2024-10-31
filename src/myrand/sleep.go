package myrand

import "time"

func RandSleep(t time.Duration) {
	time.Sleep(time.Duration(RandInt64(int64(t))))
}

func RandSleep2ms() {
	RandSleep(2 * time.Millisecond)
}
