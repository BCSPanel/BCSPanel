package httprouter

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bddjr/BCSPanel/src/server/conf"
	"github.com/bddjr/BCSPanel/src/server/mysession"
	"github.com/bddjr/BCSPanel/src/server/user"
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
	const tBase = 36
	const cookieNameBasicLoginUsed = "BCSPanelBasicLoginUsed"

	redirect := func(ctx *gin.Context) {
		scriptRedirect(ctx, 400, conf.Http.Old_PathPrefix+"login/basic/"+strconv.FormatInt(time.Now().UnixMilli(), tBase)+"/")
	}

	loginGroup.Use(handlerRemoveQuery, func(ctx *gin.Context) {
		if mysession.CheckLoggedInCookieForCtx(ctx) {
			// 已登录，脚本重定向，防止客户端丢失缓存
			scriptRedirect(ctx, 401, conf.Http.Old_PathPrefix)
			ctx.Abort()
			return
		}
		if !conf.Ssl.Old_EnableSsl {
			ctx.Header("Referrer-Policy", "same-origin")
		}
	})
	loginGroup.GET("/", redirect)
	loginGroup.GET("/basic/", redirect)

	loginGroup.GET("/basic/:t/", func(ctx *gin.Context) {
		param := ctx.Param("t")

		// 如果之前登录的时候用过这个时间戳，那么忽略本次提交，重新生成。
		// 修复Firefox的bug。
		cookieBasicLoginUsed, _ := ctx.Cookie(cookieNameBasicLoginUsed)
		if cookieBasicLoginUsed == param {
			redirect(ctx)
			return
		}
		// 参数必须是有效的
		paramTimeInt, err := strconv.ParseInt(param, tBase, 64)
		if err != nil {
			redirect(ctx)
			return
		}
		// 检查cookie时间，防止复用更旧的地址
		// 修复Firefox的bug。
		if cookieBasicLoginUsed != "" {
			cookieTimeInt, err := strconv.ParseInt(cookieBasicLoginUsed, tBase, 64)
			if err == nil && cookieTimeInt > paramTimeInt {
				redirect(ctx)
				return
			}
		}
		// 时间不能超过当前时间
		paramTime := time.UnixMilli(paramTimeInt)
		if paramTime.After(time.Now()) {
			redirect(ctx)
			return
		}

		// 使用secure
		secure := conf.Ssl.Old_EnableSsl
		if !secure {
			// 检查Referer
			referer := ctx.Request.Header.Get("Referer")
			if referer == "" {
				if paramTime.Add(3 * time.Second).After(time.Now()) {
					// 参数时间戳对比当前时间戳，相差不超过2秒
					// 浏览器不支持Referer
					ctx.String(400, "Missing Referer Header")
					return
				}
				redirect(ctx)
				return
			}
			// 判断https
			secure = strings.HasPrefix(referer, "https")
		}

		// 获取提交内容
		username, password, ok := ctx.Request.BasicAuth()
		if !ok {
			// 未提交
			ctx.Header("WWW-Authenticate", `Basic realm=`+ctx.Request.URL.Path+`, charset="UTF-8"`)
			ctx.Status(401)
			return
		}
		// 登录
		ctx.SetCookie(cookieNameBasicLoginUsed, param, 0, conf.Http.Old_PathPrefix+"login/", "", secure, true)
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
