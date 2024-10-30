package httprouter

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bddjr/BCSPanel/src/conf"
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
	if !mysession.CheckLoggedInCookieForCtx(ctx) {
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
	// 设置配置文件
	conf.Http.Old_PathPrefix = conf.Http.New_PathPrefix
	conf.Http.Old_GzipLevel = conf.Http.New_GzipLevel
	conf.Http.Old_GzipMinContentLength = conf.Http.New_GzipMinContentLength
	conf.Http.Old_EnableH2c = conf.Http.New_EnableH2c

	// 创建新的路由
	Router := gin.New()
	Router.HandleMethodNotAllowed = true
	Router.RedirectFixedPath = true

	// https://github.com/gin-gonic/gin/pull/1398
	Router.UseH2C = true

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

	const dist = "frontend-antd/dist/"

	// 404
	noRoute := func(ctx *gin.Context) {
		f, _ := os.ReadFile(dist + "404.html")
		ctx.Data(404, gin.MIMEHTML, f)
	}
	Router.NoRoute(noRoute)
	gzipstatic.NoRoute = noRoute

	// robots.txt
	Router.StaticFile("/robots.txt", dist+"robots.txt")

	// group
	mainGroup := &Router.RouterGroup
	if conf.Http.Old_PathPrefix != "/" {
		mainGroup = mainGroup.Group(conf.Http.Old_PathPrefix)
	}

	// frontend
	mainGroup.GET("/", func(ctx *gin.Context) {
		if !mysession.CheckLoggedInCookieForCtx(ctx) {
			// 未登录
			redirect(ctx, 303, "./login/")
			return
		}
		// 网页
		// ctx.File(dist)
		gzipstatic.File(ctx, dist)
	})
	files, err := os.ReadDir(dist)
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		name := f.Name()
		if name[0] == '[' {
			continue
		}
		switch name {
		case "index.html", "robots.txt", "login", "api":
			continue
		}
		if f.IsDir() {
			g := mainGroup.Group(name)
			if name == "assets" {
				g.Use(handlerCheckNotLoggedIn401)
			}
			// g.Static("/", dist+name)
			gzipstatic.Static(g, "/", dist+name)
			continue
		}
		// mainGroup.StaticFile(name, dist+name)
		gzipstatic.StaticFile(mainGroup, name, dist+name)
	}

	// login
	{
		const dist = "frontend-login2/dist/"
		g := mainGroup.Group("login")
		indexPath := g.BasePath() + "/"
		g.Use(func(ctx *gin.Context) {
			if ctx.Request.URL.Path == indexPath && mysession.CheckLoggedInCookieForCtx(ctx) {
				// 已登录
				redirect(ctx, 303, "../")
				ctx.Abort()
			}
		})
		// g.Static("/", dist)
		gzipstatic.Static(g, "/", dist)
	}

	// api
	apiInit(mainGroup.Group("api"))

	return Router.Handler()
}
