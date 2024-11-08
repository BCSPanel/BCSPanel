package router

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
	g.GET("/keepsession", a.handlerKeepSession)
}

func (a apiLogin) handlerLogin(ctx *gin.Context) {
	mysession.LogOutForCtx(ctx)

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
	u, err := user.GetForLogin(form.Username)
	if err != nil || !u.PasswordEqual(form.Password) {
		ctx.String(401, "@invalid-username-or-password")
		return
	}

	cookie, err := mysession.Create(form.Username, form.Secure)
	if err != nil {
		ctx.AbortWithError(500, err)
		return
	}
	http.SetCookie(ctx.Writer, cookie)
	ctx.Status(200)
}

func (a apiLogin) handlerLogout(ctx *gin.Context) {
	mysession.LogOutForCtx(ctx)
	redirect(ctx, 303, config.OldHttp.PathPrefix)
}

func (a apiLogin) handlerKeepSession(ctx *gin.Context) {
	if mysession.CheckCtx(ctx) {
		ctx.Status(200)
		return
	}
	ctx.Status(401)
}
