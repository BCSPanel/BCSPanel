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
	ctx.Header("Set-Cookie", cookie.String())
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
	loginGroup.GET("/", func(ctx *gin.Context) {
		// 已登录，重定向
		if mysession.CheckLoggedInCookieForCtx(ctx) {
			ctx.Redirect(303, "../")
			return
		}

		// 使用secure
		secure := conf.Ssl.Old_EnableSsl
		if !secure {
			// 检查Referer
			const qName = "t"
			const qBase = 36
			referer := ctx.Request.Header.Get("Referer")
			if referer == "" {
				// 缺少Referer
				if q := ctx.Query(qName); q != "" {
					if t, err := strconv.ParseInt(q, qBase, 64); err == nil {
						if time.Unix(t, 0).Add(10 * time.Second).After(time.Now()) {
							// 参数时间戳对比当前时间戳，相差不超过10秒
							// 浏览器不支持Referer
							ctx.String(400, "Missing Referer Header")
							return
						}
					}
				}
				// 刷新后获取Referer
				scriptRedirect(ctx, 400, "?"+qName+"="+strconv.FormatInt(time.Now().Unix(), qBase))
				return
			}
			// 判断https
			secure = strings.HasPrefix(referer, "https")
		}

		// 获取提交内容
		username, password, ok := ctx.Request.BasicAuth()
		if !ok {
			// 未提交
			ctx.Header("WWW-Authenticate", `Basic charset="UTF-8"`)
			ctx.Status(401)
			return
		}
		// 登录
		cookie, err := user.Login(username, password, secure)
		if err != nil {
			// 失败
			ctx.String(401, err.Error())
			ctx.Error(err)
			return
		}
		// 成功
		ctx.Header("Set-Cookie", cookie.String())
		scriptRedirect(ctx, 401, "../") // 防止客户端再次发送Authenticate
	})
}
