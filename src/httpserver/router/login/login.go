// 该模块仅负责登录、登出、保持登录。
// 获取用户信息应当写在users模块。

package login

import (
	"net/http"

	"github.com/bddjr/BCSPanel/src/httpserver/router/routertools"
	"github.com/bddjr/BCSPanel/src/mysession"
	"github.com/bddjr/BCSPanel/src/user"
	"github.com/gin-gonic/gin"
)

func Init(g *gin.RouterGroup) {
	g.POST("login", login)
	g.GET("logout", logout)
	g.GET("keepsession", keepSession)
}

func login(ctx *gin.Context) {
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
	if !user.Login(form.Username, form.Password) {
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

func logout(ctx *gin.Context) {
	mysession.LogOutForCtx(ctx)
	routertools.Redirect(ctx, 303, "../../")
}

func keepSession(ctx *gin.Context) {
	if mysession.CheckCtx(ctx) {
		ctx.Status(200)
		return
	}
	ctx.Status(401)
}
