package mysession

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bddjr/BCSPanel/src/config"
	"github.com/bddjr/BCSPanel/src/myrand"
	"github.com/gin-gonic/gin"
)

var lock sync.Mutex

const SessionCookieName = "BCSPanelLoginSession"

const CookieUserLength = 32
const CookieSessionLength = 256
const CookieLength = CookieUserLength + CookieSessionLength

// 会话超过2小时未使用就失效
const MaxAge = 2 * time.Hour

type SessionType struct {
	UserId        int
	UserRandomId  string
	SessionId     string
	CreateTime    time.Time
	LastUsageTime time.Time
	Cookie        *http.Cookie
}

type UserSessionsType struct {
	Sessions     map[string]*SessionType // map [SessionId] *SessionType
	UserId       int
	UserRandomId string
}

type globalUsersType struct {
	UsersFromId       map[int]*UserSessionsType    // map [UserId] *UserSessions
	UsersFromRandomId map[string]*UserSessionsType // map [UserRandomId] *UserSessions
}

var globalLoggedInUsers = globalUsersType{
	UsersFromId:       map[int]*UserSessionsType{},
	UsersFromRandomId: map[string]*UserSessionsType{},
}

// 更新最后使用时间
func (s *SessionType) UpdateLastUsageTime() {
	s.LastUsageTime = time.Now()
}

func CreateLoggedInCookie(userId int, secure bool) (cookie *http.Cookie, err error) {
	lock.Lock()
	defer lock.Unlock()
	var userRandomId string
	// 尝试获取已创建的用户随机ID
	userSessions, ok := globalLoggedInUsers.UsersFromId[userId]
	if ok {
		// 找到了，引用
		userRandomId = userSessions.UserRandomId
	} else {
		// 找不到，新建一个
		i := 0
		for {
			userRandomId = myrand.RandStr64(CookieUserLength)
			if _, ok := globalLoggedInUsers.UsersFromRandomId[userRandomId]; !ok {
				break
			}
			i++
			// 防止死循环
			if i > 1000 {
				err = fmt.Errorf("can not create userRamdomId, Too many loops")
				return
			}
		}
		userSessions = &UserSessionsType{
			Sessions:     map[string]*SessionType{},
			UserId:       userId,
			UserRandomId: userRandomId,
		}
		globalLoggedInUsers.UsersFromRandomId[userRandomId] = userSessions
		globalLoggedInUsers.UsersFromId[userId] = userSessions
	}
	// 新建会话
	var sessionId string
	i := 0
	for {
		sessionId = myrand.RandStr64(CookieSessionLength)
		if _, ok = userSessions.Sessions[sessionId]; !ok {
			break
		}
		i++
		// 防止死循环
		if i > 1000 {
			err = fmt.Errorf("can not create userRamdomId, Too many loops")
			return
		}
	}
	createTime := time.Now()
	cookie = &http.Cookie{
		Name:     SessionCookieName,
		Value:    userRandomId + sessionId,
		Path:     config.OldHttp.PathPrefix,
		Secure:   secure,
		HttpOnly: true,
		// MaxAge:   int(MaxAge.Seconds()),
	}
	userSessions.Sessions[sessionId] = &SessionType{
		UserId:        userId,
		UserRandomId:  userRandomId,
		SessionId:     sessionId,
		CreateTime:    createTime,
		LastUsageTime: createTime,
		Cookie:        cookie,
	}
	return
}

func getLoggedInSession(value string) (session *SessionType, ok bool) {
	if len(value) != CookieLength {
		return nil, false
	}

	userRandomId := value[:CookieUserLength]
	sessionId := value[CookieUserLength:]

	// 获取用户的所有session
	userSessions, ok := globalLoggedInUsers.UsersFromRandomId[userRandomId]
	if !ok {
		return nil, false
	}
	// 获取指定session
	session, ok = userSessions.Sessions[sessionId]
	if !ok {
		return nil, false
	}
	// 如果session过期
	if session.LastUsageTime.Add(MaxAge).Before(time.Now()) {
		// 过期了，移除
		delete(userSessions.Sessions, session.SessionId)
		return nil, false
	}
	return session, true
}

func CheckLoggedInCookie(value string) (ok bool) {
	lock.Lock()
	defer lock.Unlock()
	_, ok = getLoggedInSession(value)
	return
}

func CheckLoggedInCookieForCtx(ctx *gin.Context) (ok bool) {
	value, err := ctx.Cookie(SessionCookieName)
	if err != nil {
		return
	}
	lock.Lock()
	defer lock.Unlock()
	session, ok := getLoggedInSession(value)
	if !ok {
		ctx.SetCookie(SessionCookieName, "x", -1, "", "", false, true)
		return
	}
	session.UpdateLastUsageTime()
	// ctx.Writer.Header().Set("Set-Cookie", session.Cookie.String())
	return true
}

func logOutCookie(value string) (session *SessionType, cookie *http.Cookie, ok bool) {
	if len(value) != CookieLength {
		return nil, nil, false
	}

	userRandomId := value[:CookieUserLength]
	sessionId := value[CookieUserLength:]

	// 查找用户的所有会话
	userSessions, ok := globalLoggedInUsers.UsersFromRandomId[userRandomId]
	if !ok {
		return nil, nil, false
	}
	// 查找指定会话
	session, ok = userSessions.Sessions[sessionId]
	if !ok {
		return nil, nil, false
	}
	// 让cookie失效
	cookie = session.Cookie
	cookie.MaxAge = -1
	cookie.Value = "x"
	// 在列表里移除会话
	delete(userSessions.Sessions, sessionId)
	// 如果用户没有会话了，移除用户
	if len(userSessions.Sessions) == 0 {
		delete(globalLoggedInUsers.UsersFromId, userSessions.UserId)
		delete(globalLoggedInUsers.UsersFromRandomId, userRandomId)
	}
	return
}

func LogOutCookie(value string) (cookie *http.Cookie, ok bool) {
	lock.Lock()
	defer lock.Unlock()
	_, cookie, ok = logOutCookie(value)
	return
}

func LogOutSessionForRequest(req *http.Request) (cookie *http.Cookie, ok bool) {
	reqCookie, err := req.Cookie(SessionCookieName)
	if err != nil {
		return nil, false
	}
	return LogOutCookie(reqCookie.Value)
}

// 清除用户的所有会话
func logOutUser(userSessions *UserSessionsType) {
	delete(globalLoggedInUsers.UsersFromRandomId, userSessions.UserRandomId)
	delete(globalLoggedInUsers.UsersFromId, userSessions.UserId)
}

func logOutUserFromSession(session *SessionType) {
	delete(globalLoggedInUsers.UsersFromRandomId, session.UserRandomId)
	delete(globalLoggedInUsers.UsersFromId, session.UserId)
}

func logOutUserFromId(userId int) {
	if userSessions, ok := globalLoggedInUsers.UsersFromId[userId]; ok {
		logOutUser(userSessions)
	}
}

func LogOutUserFromId(userId int) {
	lock.Lock()
	defer lock.Unlock()
	logOutUserFromId(userId)
}

func logOutUserFromRandomId(randomId string) {
	if userSessions, ok := globalLoggedInUsers.UsersFromRandomId[randomId]; ok {
		logOutUser(userSessions)
	}
}

func LogOutUserFromRandomId(randomId string) {
	lock.Lock()
	defer lock.Unlock()
	logOutUserFromRandomId(randomId)
}

func LogOutUserForRequest(req *http.Request) (cookie *http.Cookie, ok bool) {
	reqCookie, err := req.Cookie(SessionCookieName)
	if err != nil {
		return nil, false
	}
	lock.Lock()
	defer lock.Unlock()
	session, cookie, ok := logOutCookie(reqCookie.Value)
	if ok {
		logOutUserFromSession(session)
	}
	return
}

// 清除所有会话
func LogOutAll() {
	lock.Lock()
	defer lock.Unlock()
	globalLoggedInUsers.UsersFromId = map[int]*UserSessionsType{}
	globalLoggedInUsers.UsersFromRandomId = map[string]*UserSessionsType{}
}

func GC() {
	lock.Lock()
	defer lock.Unlock()
	// 遍历所有cookie用户
	for userId, userSessions := range globalLoggedInUsers.UsersFromId {
		for sessionId, session := range userSessions.Sessions {
			// 检查过期
			if session.LastUsageTime.Add(MaxAge).Before(time.Now()) {
				// 过期了，移除
				delete(userSessions.Sessions, sessionId)
			}
		}
		// 如果用户没有会话了，移除用户
		if len(userSessions.Sessions) == 0 {
			delete(globalLoggedInUsers.UsersFromRandomId, userSessions.UserRandomId)
			delete(globalLoggedInUsers.UsersFromId, userId)
		}
	}
}

func Reload() {
	LogOutAll()
}
