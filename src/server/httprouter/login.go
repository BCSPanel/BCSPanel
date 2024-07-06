package httprouter

import (
	"net/http"

	"github.com/bddjr/BCSPanel/src/server/conf"
	"github.com/bddjr/BCSPanel/src/server/mysession"
	"github.com/bddjr/BCSPanel/src/server/user"
	"github.com/bddjr/basiclogin-gin"
	"github.com/gin-gonic/gin"
)

type apiLogin struct{}

func (a apiLogin) Init(g *gin.RouterGroup) {
	g.POST("/login", a.handlerLogin)
	g.GET("/logout", a.handlerLogout)
	g.GET("/update-last-usage-time", a.handlerUpdateLastUsageTime)
}

func (a apiLogin) InitBasic(loginGroup *gin.RouterGroup) {
	// 使用Basic登录
	basiclogin.New(loginGroup, func(ctx *gin.Context, username, password string, secure bool) {
		cookie, err := user.Login(username, password, secure)
		if a.loginSetCookie(ctx, cookie, err) {
			// 成功
			ctx.Header("Referrer-Policy", "no-referrer")
			scriptRedirect(ctx, 401, conf.Http.Old_PathPrefix)
		}
	})
}

func (a apiLogin) handlerLogin(ctx *gin.Context) {
	// 退出登录，如果有效
	mysession.LogOutSessionForRequest(ctx.Request)

	type loginJsonType struct {
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

	// 解析表单
	loginJson := &loginJsonType{}
	err := ctx.BindJSON(loginJson)
	if err != nil {
		ctx.String(400, err.Error())
		ctx.Error(err)
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

	if a.loginSetCookie(ctx, cookie, err) {
		// 成功
		ctx.Status(200)
	}
}

func (a apiLogin) loginSetCookie(ctx *gin.Context, cookie *http.Cookie, err error) (ok bool) {
	if err != nil {
		// 失败
		ctx.String(401, err.Error())
		ctx.Error(err)
		return false
	}
	// 成功
	ctx.Writer.Header().Add("Set-Cookie", cookie.String())
	return true
}

func (a apiLogin) handlerLogout(ctx *gin.Context) {
	// 退出登录
	cookie, ok := mysession.LogOutSessionForRequest(ctx.Request)
	if ok {
		// 会话有效，已退出
		ctx.Writer.Header().Add("Set-Cookie", cookie.String())
	} else {
		// 强制退出
		ctx.SetCookie(mysession.SessionCookieName, "x", -1, "", "", false, true)
	}
	// 重定向
	redirect(ctx, 303, conf.Http.Old_PathPrefix+"login/")
}

func (a apiLogin) handlerUpdateLastUsageTime(ctx *gin.Context) {
	if mysession.CheckLoggedInCookieForCtx(ctx) {
		ctx.Status(200)
		return
	}
	ctx.Status(401)
}
