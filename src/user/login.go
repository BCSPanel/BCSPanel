package user

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bddjr/BCSPanel/src/mysession"
)

func Register(name string, password string, inputVerificationCode string, secure bool) (cookie *http.Cookie, err error) {
	// 用户名与密码不能空
	if name == "" || password == "" {
		return nil, errors.New("no username or password entered")
	}
	// 用户名必须符合格式
	if !compiledRegexp_UsernameInputFormat.MatchString(name) {
		return nil, errors.New("the username format is incorrect")
	}

	lock.Lock()
	defer lock.Unlock()

	// 比较验证码
	if !RegisterVerifyCode.CodeEqualWithAutoClear(inputVerificationCode) {
		err = fmt.Errorf("@verify-code-mismatch")
		return
	}
	// 创建用户
	user, err := createAndWriteUser(name, password)
	if err != nil {
		return
	}
	// 创建会话
	return mysession.CreateLoggedInCookie(user.Id, secure)
}

func Login(name string, password string, secure bool) (cookie *http.Cookie, err error) {
	// 用户名或密码不能为空
	if name == "" || password == "" {
		return nil, errors.New("no username or password entered")
	}
	// 用户名必须符合格式
	if !compiledRegexp_UsernameInputFormat.MatchString(name) {
		err = errors.New("the username format is incorrect")
		return
	}

	lock.Lock()
	defer lock.Unlock()

	// 获取用户信息
	user, err := getUserFromName(name)
	// 比较密码
	if err != nil || !user.passwordEqual(password) {
		return nil, fmt.Errorf("@invalid-username-or-password")
	}
	// 更新最近登录时间
	err = user.updateAndWriteLoginTime()
	if err != nil {
		return
	}
	// 创建会话
	return mysession.CreateLoggedInCookie(user.Id, secure)
}
