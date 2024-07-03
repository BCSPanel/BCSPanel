package httprouter

import (
	"net/http"

	"github.com/bddjr/BCSPanel/src/server/conf"
	"github.com/bddjr/BCSPanel/src/server/mysession"
	"github.com/bddjr/BCSPanel/src/server/user"
	"github.com/bddjr/basiclogin-gin"
	"github.com/gin-gonic/gin"
)

type loginJson struct {
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

type apiLogin struct{}

func (a apiLogin) Init(apiGroup *gin.RouterGroup) {
	g := apiGroup.Group("login")
	g.POST("/login", a.loginHandler)
	g.GET("/logout", a.logoutHandler)
}

func (a apiLogin) loginHandler(ctx *gin.Context) {
	// 退出登录，如果有效
	mysession.LogOutSessionForRequest(ctx.Request)

	// 解析表单
	var loginJson = &loginJson{}
	err := ctx.BindJSON(loginJson)
	if err != nil {
		ctx.Status(400)
		ctx.Writer.WriteString(err.Error())
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

	if err != nil {
		// 失败
		ctx.String(401, err.Error())
		ctx.Error(err)
		return
	}
	// 成功
	ctx.Writer.Header().Add("Set-Cookie", cookie.String())
	ctx.Status(200)
}

func (a apiLogin) logoutHandler(ctx *gin.Context) {
	// 退出登录
	cookie, ok := mysession.LogOutSessionForRequest(ctx.Request)
	if ok {
		// 会话有效，已退出
		ctx.Header("Set-Cookie", cookie.String())
	}
	// 返回303
	ctx.Redirect(303, conf.Http.Old_PathPrefix+"login/")
}

func (a apiLogin) InitBasic(loginGroup *gin.RouterGroup) {
	basiclogin.New(loginGroup, func(ctx *gin.Context, username, password string, secure bool) {
		cookie, err := user.Login(username, password, secure)
		if err != nil {
			// 失败
			ctx.String(401, err.Error())
			ctx.Error(err)
			return
		}
		// 成功
		ctx.Writer.Header().Add("Set-Cookie", cookie.String())
		ctx.Header("Referrer-Policy", "no-referrer")
		scriptRedirect(ctx, 200, conf.Http.Old_PathPrefix)
	})
}
