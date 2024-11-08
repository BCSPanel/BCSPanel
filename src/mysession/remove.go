package mysession

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func removeSession(session *Session) {
	if session == nil {
		return
	}
	nameToId, ok := userNameToSessionId[session.UserName]
	if ok {
		delete(nameToId, session.sessionId)
		if len(nameToId) == 0 {
			delete(userNameToSessionId, session.UserName)
		}
	}
	delete(sessions, session.sessionId)
}

func removeCookie(w http.ResponseWriter) {
	w.Header().Add("Set-Cookie", SessionCookieName+`=0; Max-Age=0`)
}

func LogOutForCtx(ctx *gin.Context) {
	lock.Lock()
	defer lock.Unlock()

	v, err := ctx.Cookie(SessionCookieName)
	if err != nil {
		return
	}
	removeCookie(ctx.Writer)

	if len(v) != CookieValueLength {
		return
	}

	// 查找指定会话
	session, ok := get(v)
	if !ok {
		return
	}
	// 在列表里移除会话
	removeSession(session)
}

// 清除用户的所有会话
func LogOutUserForCtx(ctx *gin.Context) {
	v, err := ctx.Cookie(SessionCookieName)
	if err != nil {
		return
	}

	lock.Lock()
	defer lock.Unlock()

	s, ok := get(v)
	if !ok {
		return
	}
	nameToId, ok := userNameToSessionId[s.UserName]
	if !ok {
		return
	}

	for id := range nameToId {
		delete(sessions, id)
	}
	delete(userNameToSessionId, s.UserName)
	removeCookie(ctx.Writer)
}

// 清除所有会话
func LogOutAll() {
	lock.Lock()
	defer lock.Unlock()
	userNameToSessionId = make(map[string]map[id]struct{})
	sessions = make(map[id]*Session)
}
