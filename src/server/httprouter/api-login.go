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
