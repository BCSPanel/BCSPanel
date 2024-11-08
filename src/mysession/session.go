package mysession

import (
	"sync"
	"time"
)

var lock sync.Mutex

// map[id]*Session
var sessions = make(map[id]*Session)

// map[name]map[id]struct{}
var userNameToSessionId = make(map[string]map[id]struct{})

const SessionCookieName = "BCSPanel"

// 会话超过24小时未使用就失效
const MaxAge = 24 * time.Hour

const SessionIdLength = 16
const SessionPasswordLength = 24
const CookieValueLength = SessionIdLength + SessionPasswordLength

type id [SessionIdLength]byte

func toId(id string) (out id) {
	copy(out[:], id)
	return
}

func (id id) String() string {
	return string(id[:])
}

type passwd [SessionPasswordLength]byte

func toPasswd(passwd string) (out passwd) {
	copy(out[:], passwd)
	return
}

func (passwd passwd) String() string {
	return string(passwd[:])
}

type Session struct {
	sessionId       id
	sessionPassword passwd
	UserName        string
	CreateTime      time.Time
	LastUsageTime   time.Time
}

// 更新最后使用时间
func (s *Session) UpdateLastUsageTime() {
	s.LastUsageTime = time.Now()
}
