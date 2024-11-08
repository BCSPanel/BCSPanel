package mysession

import (
	"crypto/hmac"
	"time"

	"github.com/gin-gonic/gin"
)

func get(value string) (session *Session, ok bool) {
	if len(value) != CookieValueLength {
		return nil, false
	}

	id := value[:SessionIdLength]
	passwd := value[SessionIdLength:]

	session, ok = sessions[toId(id)]
	if !ok {
		return nil, false
	}
	// 如果session过期了，移除
	if session.LastUsageTime.Add(MaxAge).Before(time.Now()) {
		removeSession(session)
		return nil, false
	}
	// check password, anti timing attacks
	if !hmac.Equal([]byte(passwd), session.sessionPassword[:]) {
		return nil, false
	}
	return session, true
}

func CheckCtx(ctx *gin.Context) bool {
	value, err := ctx.Cookie(SessionCookieName)
	if err != nil {
		return false
	}
	lock.Lock()
	defer lock.Unlock()
	session, ok := get(value)
	if ok {
		session.UpdateLastUsageTime()
	} else {
		removeCookie(ctx.Writer)
	}
	return ok
}
