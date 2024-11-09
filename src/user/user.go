package user

import (
	"errors"
	"sync"
	"time"

	"github.com/bddjr/BCSPanel/src/mylog"
	"golang.org/x/crypto/bcrypt"
)

var publicFuncLock sync.Mutex

type User struct {
	// RegExp ^[\w\-]{1,32}$
	Name string `json:"name"`

	// Unix
	RecentLoginTime int64 `json:"recentLoginTime"`

	// Unix
	CreateTime int64 `json:"createTime"`

	// bcrypt
	Password string `json:"password"`
}

func (u *User) passwordSet(inputPassword string) error {
	b, err := bcrypt.GenerateFromPassword([]byte(inputPassword), bcrypt.DefaultCost)
	if err == nil {
		u.Password = string(b)
	}
	return err
}

func (u *User) PasswordSet(inputPassword string) error {
	publicFuncLock.Lock()
	defer publicFuncLock.Unlock()
	return u.passwordSet(inputPassword)
}

func (u *User) passwordEqual(inputPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(inputPassword))
	return err == nil
}

func (u *User) PasswordEqual(inputPassword string) bool {
	publicFuncLock.Lock()
	defer publicFuncLock.Unlock()
	return u.passwordEqual(inputPassword)
}

func (u *User) recentLoginTimeSet() {
	u.RecentLoginTime = time.Now().Unix()
}

func (u *User) write() error {
	return writeUserToSysFile(u)
}

func Create(name, password string) (*User, error) {
	if !RegexpUsernameFormat.MatchString(name) {
		return nil, errors.New("user: create error: name does not conform format")
	}

	publicFuncLock.Lock()
	defer publicFuncLock.Unlock()

	if exist(name) {
		return nil, errors.New("user: create error: name exist")
	}

	mylog.INFO("user: create ", name)

	u := &User{
		Name:       name,
		CreateTime: time.Now().Unix(),
	}
	err := u.passwordSet(password)
	if err == nil {
		err = u.write()
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func Get(name string) (*User, error) {
	if !RegexpUsernameFormat.MatchString(name) {
		return nil, errors.New("user: get error: name does not conform format")
	}

	publicFuncLock.Lock()
	defer publicFuncLock.Unlock()
	c, ok := cache.get(name)
	if ok {
		c.updateRequestTime()
		return c.user, nil
	}
	u, err := readUserFromSysFile(name)
	if err == nil {
		cache.add(u)
	}
	return u, err
}

func Login(name string, passwd string) bool {
	u, err := Get(name)
	if err != nil || !u.passwordEqual(passwd) {
		return false
	}
	u.recentLoginTimeSet()
	u.write()
	return true
}

func exist(name string) bool {
	c, ok := cache.get(name)
	if ok {
		c.updateRequestTime()
		return true
	}
	return userSysFileExist(name)
}

func Exist(name string) bool {
	if !RegexpUsernameFormat.MatchString(name) {
		return false
	}

	publicFuncLock.Lock()
	defer publicFuncLock.Unlock()
	return exist(name)
}

func Remove(name string) error {
	if !RegexpUsernameFormat.MatchString(name) {
		return errors.New("user: remove error: name does not conform format")
	}
	publicFuncLock.Lock()
	defer publicFuncLock.Unlock()
	cache.del(name)
	return removeUserSysFile(name)
}
