package user

import "time"

const cacheMaxAge = 10 * time.Second

var cache = cacheMap{}

type cacheMap map[string]*cacheItem

func (c cacheMap) add(u *User) {
	c[u.Name] = &cacheItem{
		user:              u,
		recentRequestTime: time.Now(),
	}
}

func (c cacheMap) get(name string) (*cacheItem, bool) {
	out, ok := c[name]
	return out, ok
}

func (c cacheMap) del(name string) {
	delete(c, name)
}

func (c *cacheMap) clear() {
	*c = cacheMap{}
}

type cacheItem struct {
	user              *User
	recentRequestTime time.Time
}

func (c cacheItem) updateRequestTime() {
	c.recentRequestTime = time.Now()
}
