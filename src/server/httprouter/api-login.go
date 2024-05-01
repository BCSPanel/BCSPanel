package httprouter

import (
	"net/http"

	"github.com/bddjr/BCSPanel/src/server/conf"
	"github.com/bddjr/BCSPanel/src/server/mysession"
	"github.com/bddjr/BCSPanel/src/server/user"
	"github.com/gin-gonic/gin"
)

type LoginJson struct {
	// 安全上下文
	Secure bool `json:"secure"`
	// 是否处于注册模式
	Isregister bool `json:"isregister"`
	// 用户名
	Username string `json:"username"`
	// 密码
	Password string `json:"password"`
	// 注册模式发送验证码
	VerificationCode string `json:"verification_code"`
}

func routerApiLoginInit() {
	Router.GET(conf.Http.Old_PathPrefix+"api-login/color-scheme", colorSchemeHandler)
	Router.POST(conf.Http.Old_PathPrefix+"api-login/login", loginHandler)
	pathLogout := conf.Http.Old_PathPrefix + "api-login/logout"
	Router.POST(pathLogout, logoutHandler)
	Router.GET(pathLogout, logoutHandler)
}

func loginHandler(ctx *gin.Context) {
	// 退出登录，如果有效
	mysession.LogOutSessionForRequest(ctx.Request)

	// 解析表单
	var loginJson = &LoginJson{}
	err := ctx.BindJSON(loginJson)
	if err != nil {
		ctx.AbortWithError(400, err)
		return
	}

	var cookie *http.Cookie
	if loginJson.Isregister {
		// 注册
		cookie, err = user.Register(loginJson.Username, loginJson.Password, loginJson.VerificationCode, loginJson.Secure)
	} else {
		// 登录
		cookie, err = user.Login(loginJson.Username, loginJson.Password, loginJson.Secure)
	}

	if err != nil {
		// 失败
		ctx.Status(401)
		ctx.Writer.WriteString(err.Error())
		ctx.Abort()
		return
	}
	// 成功
	ctx.Writer.Header().Set("Set-Cookie", cookie.String())
	ctx.AbortWithStatus(200)
}

func loginSimpleHandler(ctx *gin.Context) {
	type objT struct {
		ColorScheme string
		IsReg       bool
		Err         string
	}
	obj := &objT{
		ColorScheme: ColorScheme,
		IsReg:       false,
		Err:         "",
	}
	if obj.ColorScheme == "" {
		obj.ColorScheme = "light dark"
	}

	// 使用GET请求时触发
	if ctx.Request.Method != "POST" {
		// ctx.Writer.Header().Add("Last-Modified", loginSimpleFileLastModified)
		ctx.Writer.Header().Add("Etag", loginSimpleFileEtag)
		// if loginSimpleFileLastModified != "" && ctx.Request.Header.Get("If-Modified-Since") == loginSimpleFileLastModified {
		if loginSimpleFileEtag != "" {
			noneMatch := ctx.Request.Header.Get("If-None-Match")
			if noneMatch == "" {
				noneMatch = ctx.Request.Header.Get("If-Match")
			}
			if noneMatch == loginSimpleFileEtag {
				// 缓存有效
				ctx.AbortWithStatus(304)
				return
			}
		}
		// 返回html
		ctx.HTML(200, loginSimpleFileName, obj)
		return
	}

	// 退出前端已登录的会话，如果有
	mysession.LogOutSessionForRequest(ctx.Request)

	// 解析提交的表单
	form := LoginJson{
		Secure:           ctx.PostForm("secure") == "on",
		Isregister:       ctx.PostForm("isregister") == "on",
		Username:         ctx.PostForm("username"),
		Password:         ctx.PostForm("password"),
		VerificationCode: ctx.PostForm("verification_code"),
	}

	var cookie *http.Cookie
	var err error

	if form.Isregister {
		// 注册模式
		obj.IsReg = true
		if form.Password != ctx.PostForm("repeat_password") {
			// 输入的两个密码不一致（不涉及安全问题，无需使用hmac标准库比较）
			obj.Err = "Don't enter two different passwords"
			ctx.HTML(401, loginSimpleFileName, obj)
			return
		}
		cookie, err = user.Register(form.Username, form.Password, form.VerificationCode, form.Secure)
	} else {
		// 登录模式
		cookie, err = user.Login(form.Username, form.Password, form.Secure)
	}

	if err != nil {
		// 失败
		obj.Err = err.Error()
		ctx.HTML(401, loginSimpleFileName, obj)
		return
	}
	// 成功
	ctx.Writer.Header().Set("Set-Cookie", cookie.String())
	ctx.Redirect(303, conf.Http.Old_PathPrefix)
}

func logoutHandler(ctx *gin.Context) {
	// 退出登录
	cookie, ok := mysession.LogOutSessionForRequest(ctx.Request)
	if ok {
		// 会话有效，已退出
		ctx.Writer.Header().Set("Set-Cookie", cookie.String())
	}
	// 返回303
	ctx.Redirect(303, conf.Http.Old_PathPrefix+"login/")
}
