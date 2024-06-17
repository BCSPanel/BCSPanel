package httprouter

import (
	"fmt"
	"strings"

	"github.com/bddjr/BCSPanel/src/server/conf"
	"github.com/bddjr/BCSPanel/src/server/mylog"
	"github.com/bddjr/BCSPanel/src/server/mysession"
	"github.com/gin-gonic/gin"
)

var Router *gin.Engine
var GzipHandler gin.HandlerFunc

func Init() {
	// UpdateRouter()
}

func UpdateRouter() {
	// 设置配置文件
	conf.Http.Old_PathPrefix = conf.Http.New_PathPrefix

	// 日志
	mylog.INFOln("httprouter UpdateRouter")

	// 更新前置
	UpdateGzipHandler()
	UPdate404html()
	UpdateColorScheme()
	updateRegexp()

	// 创建新的路由
	Router = gin.New()
	// 会返回不支持的方法
	Router.HandleMethodNotAllowed = true
	// 中间件
	Router.Use(
		GinLoggerHandler,
		myRouter,
	)
	if GzipHandler != nil {
		Router.Use(GzipHandler)
	}
	// 404
	Router.NoRoute(_404Handler)

	// robots.txt
	Router.GET("/robots.txt", RobotsTxtHandler)
	Router.HEAD("/robots.txt", RobotsTxtHandler)

	// login
	if conf.Http.Old_PathPrefix == "/" {
		r := func(ctx *gin.Context) {
			ctx.Redirect(302, "/login/icon/BCSP-64x64.png")
		}
		Router.GET("/favicon.ico", r)
		Router.HEAD("/favicon.ico", r)
	}
	p := conf.Http.Old_PathPrefix + "login/"
	Router.Static(p, "./src/web-login/dist")
	Router.POST(p, func(ctx *gin.Context) {
		ctx.Redirect(303, p)
	})

	// web
	Router.StaticFile(conf.Http.Old_PathPrefix, "./src/web/dist/index.html")
	Router.Static(conf.Http.Old_PathPrefix+"assets/", "./src/web/dist/assets/")
	Router.Static(conf.Http.Old_PathPrefix+"icon/", "./src/web/dist/icon/")

	// ok
	Router.GET(conf.Http.Old_PathPrefix+"ok", func(ctx *gin.Context) {
		ctx.String(200, "ok")
	})

	// api-login
	routerApiLoginInit()
}

func RobotsTxtHandler(ctx *gin.Context) {
	if !conf.Robots.EnableRobotsTxt {
		_404Handler(ctx)
		return
	}
	ctx.Data(200, gin.MIMEPlain, []byte("User-agent: *\nDisallow: /\n"))
}

func myRouter(ctx *gin.Context) {
	// fmt.Println("myRouter")

	WH := ctx.Writer.Header()
	Hostname := RequestHostname(ctx.Request)
	// 如果与已允许的 Host 不匹配，返回421
	if conf.Ssl_NeedToReturn421ForUnknownServername(Hostname) {
		ctx.AbortWithStatus(421)
		return
	}
	// 标头防搜索引擎
	if conf.Robots.EnableXRobotsTag {
		WH.Add("X-Robots-Tag", "noindex, nofollow")
	}

	// 如果ssl开启
	if conf.Ssl.Old_EnableSsl && conf.Ssl.Only_HSTS != "" {
		// HSTS
		WH.Add("Strict-Transport-Security", conf.Ssl.Only_HSTS)
	}

	// 请求路径
	Path := ctx.Request.URL.Path

	// 路径不能有连在一起的多个斜杠
	if compiledRegExp_myRouter_moreSlash.MatchString(Path) {
		ctx.Redirect(301, compiledRegExp_myRouter_moreSlash.ReplaceAllString(Path, "/"))
		ctx.Abort()
		return
	}

	// 如果路径开头错误
	if !strings.HasPrefix(Path, conf.Http.Old_PathPrefix) {
		WH.Add("Cache-Control", "no-cache")
		return
	}

	// 裁剪 /bcspanel/ => /
	Path = Path[len(conf.Http.Old_PathPrefix)-1:]

	// 涉及到某些目录时，检查用户有没有登录
	if Path == "/" {
		// 未登录，网页重定向到login
		if !mysession.CheckLoggedInCookieForCtx(ctx) {
			ctx.Redirect(303, "./login/")
			ctx.Abort()
			return
		}
	} else if Path == "/login/" {
		// 已登录，重定向到/
		if mysession.CheckLoggedInCookieForCtx(ctx) {
			ctx.Redirect(303, "../")
			ctx.Abort()
			return
		}
	} else if compiledRegExp_myRouter_notLoggedIn401.MatchString(Path) {
		// 未登录，不能访问私密api或私密文件
		if !mysession.CheckLoggedInCookieForCtx(ctx) {
			WH.Del("Cache-Control")
			WH.Add("Refresh", fmt.Sprint("0; URL=", conf.Http.Old_PathPrefix, "login/"))
			ctx.Data(401, gin.MIMEHTML, []byte("401 Unauthorized\n"))
			ctx.Abort()
			return
		}
	}

	// 缓存
	// 必须向服务器确认有效
	WH.Add("Cache-Control", "no-cache")

	// // 继续执行别的操作，完成了再回来
	// ctx.Next()

	// // 似乎发生了错误，尝试移除缓存标头。
	// // 只要没开始写正文就能移除，移除不了也不会报错。
	// if ctx.Writer.Status() >= 400 {
	// 	WH.Del("Cache-Control")
	// }
}
