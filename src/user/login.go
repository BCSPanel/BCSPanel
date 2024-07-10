package user

import (
	"fmt"
	"net/http"

	"github.com/bddjr/BCSPanel/src/myregexp"
	"github.com/bddjr/BCSPanel/src/mysession"
)

func Register(name string, password string, inputVerificationCode string, secure bool) (cookie *http.Cookie, err error) {
	// 用户名与密码不能空
	if name == "" || password == "" {
		return nil, fmt.Errorf("user_name_or_password_not_filled_in")
	}
	// 用户名必须符合格式
	if !compiledRegexp_UsernameInputFormat.MatchString(name) {
		err = fmt.Errorf("incorrect_username_format")
		return
	}
	// 密码不能太长
	if len(password) > MaxPasswordLength {
		err = fmt.Errorf("password_too_long")
		return
	}
	// 判断密码强度必须够
	if len(password) < 12 ||
		!myregexp.Compiled_az.MatchString(password) ||
		!myregexp.Compiled_AZ.MatchString(password) ||
		!myregexp.Compiled_09.MatchString(password) {
		err = fmt.Errorf("password_strength_is_insufficient")
		return
	}

	lock.Lock()
	defer lock.Unlock()

	// 比较验证码
	if !RegisterVerifyCode.CodeEqualWithAutoClear(inputVerificationCode) {
		err = fmt.Errorf("wrong_verification_code")
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
		err = fmt.Errorf("user_name_or_password_not_filled_in")
		return
	}
	// 用户名必须符合格式
	if !compiledRegexp_UsernameInputFormat.MatchString(name) {
		err = fmt.Errorf("incorrect_username_format")
		return
	}
	// 密码不能太长
	if len(password) > MaxPasswordLength {
		return nil, fmt.Errorf("password_too_long")
	}
	lock.Lock()
	defer lock.Unlock()
	// 获取用户信息
	user, err := getUserFromName(name)
	// 比较密码
	if err != nil || !user.passwordEqual(password) {
		err = fmt.Errorf("wrong_username_or_password")
		return
	}
	// 更新最近登录时间
	err = user.updateAndWriteLoginTime()
	if err != nil {
		return
	}
	// 创建会话
	return mysession.CreateLoggedInCookie(user.Id, secure)
}
