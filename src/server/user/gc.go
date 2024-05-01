package user

import "time"

func GC() {
	lock.Lock()
	defer lock.Unlock()
	RegisterVerifyCode.IsValidWithAutoClear()
	for id, cache := range usersCache.UserFromId {
		if cache.LastReadTime.Add(UserCacheTimeout).Before(time.Now()) {
			delete(usersCache.UserFromName, cache.User.Name)
			delete(usersCache.UserFromId, id)
		}
	}
}

func Reload() {
	lock.Lock()
	defer lock.Unlock()
	usersCache.UserFromName = map[string]*UserCacheType{}
	usersCache.UserFromId = map[int]*UserCacheType{}
}
