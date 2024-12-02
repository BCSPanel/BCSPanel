package mysession

import (
	"errors"
	"net/http"
	"time"

	"github.com/bddjr/BCSPanel/src/config"
	"github.com/bddjr/cryptorandstr"
)

func Create(username string, secure bool) (cookie *http.Cookie, err error) {
	lock.Lock()
	defer lock.Unlock()

	var idStr string
	var id id
	for i := 0; ; i++ {
		idStr = cryptorandstr.MustRand64(SessionIdLength)
		id = toId(idStr)
		if _, ok := sessions[id]; !ok {
			break
		}
		if i >= 1000 {
			return nil, errors.New("session: Unable to generate a valid ID")
		}
	}

	passwdStr := cryptorandstr.MustRand64(SessionPasswordLength)
	t := time.Now()

	sessions[id] = &Session{
		sessionId:       id,
		sessionPassword: toPasswd(passwdStr),
		UserName:        username,
		CreateTime:      t,
		LastUsageTime:   t,
	}

	cookie = &http.Cookie{
		Name:     SessionCookieName,
		Value:    idStr + passwdStr,
		Path:     config.OldHttp.PathPrefix,
		Secure:   secure,
		HttpOnly: true,
	}
	return
}
