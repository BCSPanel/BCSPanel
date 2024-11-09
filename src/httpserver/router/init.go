package router

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bddjr/BCSPanel/src/config"
	"github.com/bddjr/BCSPanel/src/httpserver/router/files"
	"github.com/bddjr/BCSPanel/src/httpserver/router/login"
	"github.com/bddjr/BCSPanel/src/httpserver/router/settings"
	"github.com/bddjr/BCSPanel/src/httpserver/router/terminals"
	"github.com/bddjr/BCSPanel/src/httpserver/router/users"
	"github.com/bddjr/BCSPanel/src/mylog"
	"github.com/bddjr/BCSPanel/src/mysession"
	"github.com/bddjr/gzipstatic-gin"
	"github.com/gin-gonic/gin"
	"github.com/nanmu42/gzip"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func api(g *gin.RouterGroup) {
	files.Init(g.Group("files"))
	login.Init(g.Group("login"))
	settings.Init(g.Group("settings"))
	terminals.Init(g.Group("terminals"))
	users.Init(g.Group("users"))
}

func GetHandler() http.Handler {
	// 创建新的路由
	Router := gin.New()
	Router.HandleMethodNotAllowed = true
	Router.RedirectFixedPath = true

	// https://github.com/gin-gonic/gin/pull/1398
	Router.UseH2C = config.OldHttp.H2C

	// 中间件
	Router.Use(
		gin.LoggerWithConfig(gin.LoggerConfig{
			Output: &mylog.Writer,
			Formatter: func(param gin.LogFormatterParams) string {
				v := fmt.Sprint(
					// 时间
					param.TimeStamp.Format("2006/01/02 15:04:05"), " [GIN] ",
					// 客户端IP
					param.ClientIP, " ",
					// 状态码
					param.StatusCode, " ",
					// 请求方法
					param.Method, " ",
					// 请求路径
					param.Path,
				)
				// 错误消息
				if param.ErrorMessage != "" {
					v += " " + param.ErrorMessage
					if v[len(v)-1] == '\n' {
						return v
					}
				}
				v += "\n"
				return v
			},
		}),
		func(ctx *gin.Context) {
			// 拒绝跨域请求
			origin := ctx.GetHeader("Origin")
			switch origin {
			case "",
				"https://" + ctx.Request.Host,
				"http://" + ctx.Request.Host:
				// 没跨域
			default:
				// 跨域了
				ctx.Writer.Header().Del("Allow")
				ctx.AbortWithError(403, fmt.Errorf("cross origin request from %q", origin))
				return
			}
		},
		func(ctx *gin.Context) {
			ctx.Header("Cache-Control", "no-cache")
			ctx.Header("X-Robots-Tag", "noindex, nofollow")
			ctx.Header("Referrer-Policy", "no-referrer")
		},
		gzip.DefaultHandler().Gin,
	)

	const dist = "frontend/dist/"

	// 404
	noRoute := func(ctx *gin.Context) {
		ctx.Writer.Header().Del("Cache-Control")
		f, _ := os.ReadFile(dist + "404.html")
		ctx.Data(404, gin.MIMEHTML, f)
	}
	Router.NoRoute(noRoute)
	gzipstatic.NoRoute = noRoute

	// robots.txt
	Router.StaticFile("/robots.txt", dist+"robots.txt")

	// group
	main := &Router.RouterGroup
	if config.OldHttp.PathPrefix != "/" {
		main = main.Group(config.OldHttp.PathPrefix)
	}

	{
		g := main.Group("assets")
		g.Use(func(ctx *gin.Context) {
			ctx.Header("Cache-Control", "max-age=86400")
		})
		gzipstatic.Static(g, "/", dist+"assets")
	}

	gzipstatic.StaticFile(main, "loading-failed.js", dist+"loading-failed.js")

	// index
	main.GET("/", func(ctx *gin.Context) {
		ctx.Request.Header.Del("If-Modified-Since")
		ctx.Header("Cache-Control", "no-store")
		if mysession.CheckCtx(ctx) {
			gzipstatic.File(ctx, dist)
		} else {
			gzipstatic.File(ctx, dist+"login.html")
		}
	})

	// api
	api(main.Group("api"))

	return Router.Handler()
}
