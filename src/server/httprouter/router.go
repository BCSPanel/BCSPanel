package httprouter

import (
	"net/http"

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

	// 创建新的路由
	Router = gin.New()
	Router.HandleMethodNotAllowed = true
	Router.RedirectFixedPath = true
	// 中间件
	Router.Use(
		GinLoggerHandler,
		mainUse,
	)
	if GzipHandler != nil {
		Router.Use(GzipHandler)
	}
	// 404
	Router.NoRoute(_404Handler)

	// robots.txt
	Router.GET("/robots.txt", robotsTxtHandler)
	Router.HEAD("/robots.txt", robotsTxtHandler)

	// web
	Router.GET(conf.Http.Old_PathPrefix, indexHtmlHandler)
	Router.HEAD(conf.Http.Old_PathPrefix, indexHtmlHandler)
	{
		// assets
		g := Router.Group(conf.Http.Old_PathPrefix + "assets/")
		g.Use(checkNotLoggedIn401)
		g.Static("/", "./src/web/dist/assets/")
	}
	{
		// icon
		g := Router.Group(conf.Http.Old_PathPrefix + "icon/")
		g.Use(checkNotLoggedIn401)
		g.Static("/", "./src/web/dist/icon/")
	}

	// login
	{
		g := Router.Group(conf.Http.Old_PathPrefix + "login/")
		g.GET("/", loginHtmlHandler)
		g.HEAD("/", loginHtmlHandler)
		g.Static("/assets/", "./src/web-login/dist/assets/")
		g.Static("/config/", "./src/web-login/dist/config/")
		g.Static("/icon/", "./src/web-login/dist/icon/")
		g.Static("/ie/", "./src/web-login/dist/ie/")
	}

	// api-login
	routerApiLoginInit()
}

func indexHtmlHandler(ctx *gin.Context) {
	if !mysession.CheckLoggedInCookieForCtx(ctx) {
		ctx.Redirect(303, "./login/")
		return
	}
	http.ServeFile(ctx.Writer, ctx.Request, "./src/web/dist/index.html")
}

func loginHtmlHandler(ctx *gin.Context) {
	if mysession.CheckLoggedInCookieForCtx(ctx) {
		ctx.Redirect(303, "../")
		return
	}
	http.ServeFile(ctx.Writer, ctx.Request, "./src/web-login/dist/index.html")
}

func robotsTxtHandler(ctx *gin.Context) {
	if !conf.Robots.EnableRobotsTxt {
		_404Handler(ctx)
		return
	}
	ctx.Data(200, gin.MIMEPlain, []byte("User-agent: *\nDisallow: /\n"))
}

func checkNotLoggedIn401(ctx *gin.Context) {
	if !mysession.CheckLoggedInCookieForCtx(ctx) {
		ctx.AbortWithStatus(401)
	}
}

func mainUse(ctx *gin.Context) {
	WH := ctx.Writer.Header()
	// 标头防搜索引擎
	if conf.Robots.EnableXRobotsTag {
		WH.Set("X-Robots-Tag", "noindex, nofollow")
	}
	// 缓存必须向服务器确认有效
	WH.Set("Cache-Control", "no-cache")
	// HSTS
	if conf.Ssl.Old_EnableSsl && conf.Ssl.Only_HSTS != "" {
		WH.Set("Strict-Transport-Security", conf.Ssl.Only_HSTS)
	}
}
