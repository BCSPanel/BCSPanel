package router

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bddjr/BCSPanel/src/config"
	"github.com/bddjr/BCSPanel/src/mylog"
	"github.com/bddjr/BCSPanel/src/mysession"
	"github.com/bddjr/gzipstatic-gin"
	"github.com/bddjr/hlfhr"
	"github.com/gin-gonic/gin"
	"github.com/nanmu42/gzip"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func handlerCheckNotLoggedIn401(ctx *gin.Context) {
	if !mysession.CheckCtx(ctx) {
		ctx.AbortWithStatus(401)
	}
}

func redirect(ctx *gin.Context, code int, path string) {
	hlfhr.Redirect(ctx.Writer, code, path)
}

func apiInit(apiGroup *gin.RouterGroup) {
	apiFiles{}.Init(apiGroup.Group("files"))
	apiLogin{}.Init(apiGroup.Group("login"))
	apiSettings{}.Init(apiGroup.Group("settings"))
	apiTerminals{}.Init(apiGroup.Group("terminals"))
	apiUsers{}.Init(apiGroup.Group("users"))
}

// var regexpUserAgentBot = regexp.MustCompile(`[bB][oO][tT]`)

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
			wh := ctx.Writer.Header()
			// 拒绝跨域请求
			origin := ctx.GetHeader("Origin")
			if len(origin) > len("https://")+255+len(":65535") {
				// 请求头内容太长了，不像正常的
				ctx.AbortWithStatus(431)
				return
			}
			switch origin {
			case "",
				"https://" + ctx.Request.Host,
				"http://" + ctx.Request.Host:
				// 没跨域
			default:
				// 跨域了
				wh.Del("Allow")
				ctx.AbortWithError(403, fmt.Errorf("cross origin request from %q", origin))
				return
			}
			wh.Set("Cache-Control", "no-cache")
			wh.Set("X-Robots-Tag", "noindex, nofollow")
			wh.Set("Referrer-Policy", "no-referrer")
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
	mainGroup := &Router.RouterGroup
	if config.OldHttp.PathPrefix != "/" {
		mainGroup = mainGroup.Group(config.OldHttp.PathPrefix)
	}

	{
		g := mainGroup.Group("assets")
		g.Use(func(ctx *gin.Context) {
			ctx.Header("Cache-Control", "max-age=86400")
		})
		gzipstatic.Static(g, "/", dist+"assets")
	}

	gzipstatic.StaticFile(mainGroup, "loading-failed.js", dist+"loading-failed.js")

	// index
	mainGroup.GET("/", func(ctx *gin.Context) {
		ctx.Request.Header.Del("If-Modified-Since")
		ctx.Header("Cache-Control", "no-store")
		if mysession.CheckCtx(ctx) {
			gzipstatic.File(ctx, dist)
		} else {
			gzipstatic.File(ctx, dist+"login.html")
		}
	})

	// api
	apiInit(mainGroup.Group("api"))

	return Router.Handler()
}
