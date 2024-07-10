package user

import (
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bddjr/BCSPanel/src/server/myhmac"
	"github.com/bddjr/BCSPanel/src/server/myrand"
)

const TimeFormat = "2006-01-02 15:04:05 -07:00 MST"
const MaxPasswordLength = 128
const DataUsersDirPath = "data/users"
const NameToIdDirPath = DataUsersDirPath + "/#name-to-id"
const PathPrefixForUserNameToId = NameToIdDirPath + "/@"
const NextIdPath = DataUsersDirPath + "/#next-id"

const UserCacheTimeout = 1 * time.Hour

var lock sync.Mutex

type UserVersion struct {
	Version int
}

type UserV1 struct {
	Version       int    `json:"Version"`
	Name          string `json:"Name"`
	Id            int    `json:"Id"`
	RegisterTime  string `json:"RegisterTime"`
	LastLoginTime string `json:"LastLoginTime"`
	// Password      UserPasswordV1 `json:"Password"`
}

type UserCacheType struct {
	User         *UserV1
	LastReadTime time.Time
}

type UsersCacheType struct {
	UserFromId   map[int]*UserCacheType
	UserFromName map[string]*UserCacheType
}

var usersCache = UsersCacheType{
	UserFromId:   map[int]*UserCacheType{},
	UserFromName: map[string]*UserCacheType{},
}

func (user *UserV1) createUserCache() *UserCacheType {
	return &UserCacheType{
		User:         user,
		LastReadTime: time.Now(),
	}
}

func (userCache *UserCacheType) updateUserCacheReadTime() {
	userCache.LastReadTime = time.Now()
}

func userNameToId(name string) (id int, err error) {
	if !compiledRegexp_UsernameInputFormat.MatchString(name) {
		err = fmt.Errorf("userNameToId: incorrect username format")
		return
	}
	name = strings.ToLower(name)
	if userCache, ok := usersCache.UserFromName[name]; ok {
		userCache.updateUserCacheReadTime()
		return userCache.User.Id, nil
	}
	fb, err := os.ReadFile(PathPrefixForUserNameToId + name)
	if err != nil {
		return
	}
	return strconv.Atoi(string(fb))
	// return strconv.ParseUint(string(fb), 10, 64)
}

func UserNameToId(name string) (int, error) {
	lock.Lock()
	defer lock.Unlock()
	return userNameToId(name)
}

func writeUserNameToId(name string, id int) error {
	if !compiledRegexp_UsernameInputFormat.MatchString(name) {
		return fmt.Errorf("writeUserNameToId: incorrect username format")
	}
	name = strings.ToLower(name)
	os.MkdirAll(NameToIdDirPath, 0777)
	return os.WriteFile(PathPrefixForUserNameToId+name, []byte(strconv.Itoa(id)), 0777)
}

func WriteUserNameToId(name string, id int) error {
	lock.Lock()
	defer lock.Unlock()
	return writeUserNameToId(name, id)
}

func deleteUserNameToId(name string) error {
	name = strings.ToLower(name)
	delete(usersCache.UserFromName, name)
	return os.Remove(PathPrefixForUserNameToId + name)
}

func getUserFromId(id int) (user *UserV1, err error) {
	if id == 0 {
		err = fmt.Errorf("getUser: id == 0")
		return
	}
	if userCache, ok := usersCache.UserFromId[id]; ok {
		userCache.updateUserCacheReadTime()
		return userCache.User, nil
	}
	fb, err := os.ReadFile(fmt.Sprint(DataUsersDirPath, "/", id, ".json"))
	if err != nil {
		return
	}
	user = &UserV1{}
	err = json.Unmarshal(fb, user)
	if err != nil {
		return
	}
	userCache := &UserCacheType{
		User: user,
	}
	userCache.updateUserCacheReadTime()
	usersCache.UserFromId[id] = userCache
	usersCache.UserFromName[user.Name] = userCache
	return
}

func GetUserFromId(id int) (user *UserV1, err error) {
	lock.Lock()
	defer lock.Unlock()
	return getUserFromId(id)
}

func getUserFromName(name string) (user *UserV1, err error) {
	id, err := userNameToId(name)
	if err != nil {
		return
	}
	if id == 0 {
		err = fmt.Errorf("getUser: id == 0")
		return
	}
	user, err = getUserFromId(id)
	if err != nil {
		return
	}
	if user.Name != name {
		os.Remove(name)
		delete(usersCache.UserFromName, name)
		err = fmt.Errorf("getUser: username mismatch")
	}
	return
}

func GetUserFromName(name string) (user *UserV1, err error) {
	lock.Lock()
	defer lock.Unlock()
	return getUserFromName(name)
}

func (user *UserV1) writeUser() error {
	if user.Id == 0 {
		return fmt.Errorf("writeUser: id == 0")
	}
	userjson, err := json.Marshal(user)
	if err != nil {
		return err
	}
	os.MkdirAll(DataUsersDirPath, 0777)
	filename := fmt.Sprint(DataUsersDirPath, "/", user.Id, ".json")
	err = os.WriteFile(filename, userjson, 0777)
	if err != nil {
		return err
	}
	name := strings.ToLower(user.Name)
	if userCache, ok := usersCache.UserFromId[user.Id]; ok {
		if _, ok := usersCache.UserFromName[name]; !ok {
			usersCache.UserFromName[name] = userCache
		}
		return nil
	}
	userCache := user.createUserCache()
	usersCache.UserFromId[user.Id] = userCache
	usersCache.UserFromName[name] = userCache
	return nil
}

func (user *UserV1) WriteUser() error {
	lock.Lock()
	defer lock.Unlock()
	return user.writeUser()
}

func createAndWriteUser(name string, password string) (user *UserV1, err error) {
	if _, err := os.Stat(PathPrefixForUserNameToId + name); err == nil {
		return nil, fmt.Errorf("the_username_already_exists")
	}
	os.MkdirAll(DataUsersDirPath, 0777)
	var id int
	fb, err := os.ReadFile(NextIdPath)
	if err != nil {
		id = 1
	} else {
		id, err = strconv.Atoi(string(fb))
		if err != nil {
			return
		}
		if id == 0 {
			err = fmt.Errorf("createUser: id == 0")
			return
		}
	}
	if _, err := os.Stat(fmt.Sprintf(DataUsersDirPath+"/%d.json", id)); err == nil {
		return nil, fmt.Errorf("createUser: user id already exist")
	}
	timeStr := time.Now().Format(TimeFormat)
	user = &UserV1{
		Version:       1,
		Name:          name,
		Id:            id,
		RegisterTime:  timeStr,
		LastLoginTime: timeStr,
	}
	err = user.writeUser()
	if err != nil {
		return
	}
	err = user.changeAndWriteUserPassword(password)
	if err != nil {
		return
	}
	os.WriteFile(NextIdPath, []byte(strconv.Itoa(id+1)), 0777)
	err = writeUserNameToId(user.Name, id)
	return
}

func CreateAndWriteUser(name string, password string) (user *UserV1, err error) {
	lock.Lock()
	defer lock.Unlock()
	return createAndWriteUser(name, password)
}

func (user *UserV1) changeAndWriteUserPassword(inputPassword string) (err error) {
	salt := myrand.RandBytes(16)
	password := &PasswordV1{}
	copy(password.Password[:], myhmac.HmacSha3_512([]byte(inputPassword), salt))
	copy(password.Salt[:], salt)
	err = password.writeForId(user.Id)
	return
}

func (user *UserV1) ChangeAndWriteUserPassword(password string) (err error) {
	lock.Lock()
	defer lock.Unlock()
	return user.changeAndWriteUserPassword(password)
}

func (user *UserV1) changeAndWriteUserName(name string) (err error) {
	if !compiledRegexp_UsernameInputFormat.MatchString(name) {
		return fmt.Errorf("changeAndWriteUserName: incorrect username format")
	}
	name = strings.ToLower(name)
	if strings.ToLower(user.Name) == name {
		return fmt.Errorf("changeAndWriteUserName: The username has not changed")
	}
	deleteUserNameToId(user.Name)
	user.Name = name
	err = user.writeUser()
	if err != nil {
		return
	}
	return writeUserNameToId(name, user.Id)
}

func (user *UserV1) passwordEqual(password string) bool {
	storedPassword, err := getPasswordFromId(user.Id)
	if err != nil {
		return false
	}
	return hmac.Equal(storedPassword.Password[:], myhmac.HmacSha3_512([]byte(password), storedPassword.Salt[:]))
}

func (user *UserV1) PasswordEqual(password string) bool {
	lock.Lock()
	defer lock.Unlock()
	return user.passwordEqual(password)
}

func (user *UserV1) updateAndWriteLoginTime() error {
	user.LastLoginTime = time.Now().Format(TimeFormat)
	return user.writeUser()
}
