package httprouter

import (
	"net/http"

	"github.com/bddjr/BCSPanel/src/config"
	"github.com/bddjr/BCSPanel/src/mysession"
	"github.com/bddjr/BCSPanel/src/user"
	"github.com/gin-gonic/gin"
)

type apiLogin struct{}

func (a apiLogin) Init(g *gin.RouterGroup) {
	g.POST("/login", a.handlerLogin)
	g.GET("/logout", a.handlerLogout)
	g.GET("/update-last-usage-time", a.handlerUpdateLastUsageTime)
}

func (a apiLogin) handlerLogin(ctx *gin.Context) {
	// 退出登录，如果有效
	mysession.LogOutSessionForRequest(ctx.Request)

	type formType struct {
		// 安全上下文
		Secure bool `json:"secure"`
		// 用户名
		Username string `json:"username"`
		// 密码
		Password string `json:"password"`
	}

	// 解析表单
	form := &formType{}
	err := ctx.BindJSON(form)
	if err != nil {
		ctx.String(400, err.Error())
		ctx.Error(err)
		return
	}

	// 登录
	user, err := user.Get(form.Username)
	if err != nil {
		ctx.Status(500)
		ctx.Error(err)
		return
	}
	if !user.PasswordEqual(form.Password) {
		// 密码错误
		ctx.String(401, err.Error())
		ctx.Error(err)
		return
	}

	cookie, err := mysession.CreateLoggedInCookie(form.Username, form.Secure)
	ctx.SetCookie()
	ctx.Status(200)
}

func (a apiLogin) loginSetCookie(ctx *gin.Context, cookie *http.Cookie, err error) (ok bool) {
	if err != nil {

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
	redirect(ctx, 303, config.OldHttp.PathPrefix)
}

func (a apiLogin) handlerUpdateLastUsageTime(ctx *gin.Context) {
	if mysession.CheckLoggedInCookieForCtx(ctx) {
		ctx.Status(200)
		return
	}
	ctx.Status(401)
}
