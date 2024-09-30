package httprouter

import (
	"crypto/hmac"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/bddjr/BCSPanel/src/conf"
	"github.com/bddjr/BCSPanel/src/mylog"
	"github.com/bddjr/BCSPanel/src/mysession"
	"github.com/bddjr/gzipstatic-gin"
	"github.com/bddjr/hlfhr"
	"github.com/gin-gonic/gin"
	"github.com/nanmu42/gzip"
)

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
	Router.UseH2C = conf.Http.Old_EnableH2c && !conf.Ssl.Old_EnableSsl

	// 中间件
	Router.Use(
		func(ctx *gin.Context) {
			if conf.Http.Only_EnableGinLog {
				handlerGinLogger(ctx)
			}
		},
		func(ctx *gin.Context) {
			wh := ctx.Writer.Header()
			// 拒绝跨域请求
			if h := ctx.Request.Header.Get("Origin"); h != "" {
				if h == "null" {
					wh.Del("Allow")
					ctx.AbortWithError(403, fmt.Errorf("cross origin request from null"))
					return
				}
				origin, err := url.Parse(h)
				if err != nil {
					wh.Del("Allow")
					ctx.AbortWithStatus(400)
					return
				}
				if origin.Host != ctx.Request.Host {
					wh.Del("Allow")
					ctx.AbortWithError(403, fmt.Errorf("cross origin request from %s", origin.Host))
					return
				}
			}
			// 增加响应头
			for _, v := range conf.Http.Only_AddHeaders {
				for k, v := range v {
					wh.Add(k, v)
				}
			}
			wh.Set("Cache-Control", "no-cache")
			wh.Set("Referrer-Policy", "no-referrer")
		},
	)

	// GZIP
	if conf.Http.Old_GzipLevel != 0 {
		Router.Use(gzip.NewHandler(gzip.Config{
			// gzip compression level to use
			CompressionLevel: conf.Http.Old_GzipLevel,
			// minimum content length to trigger gzip, the unit is in byte.
			MinContentLength: conf.Http.Old_GzipMinContentLength,
			// RequestFilter decide whether or not to compress response judging by request.
			// Filters are applied in the sequence here.
			RequestFilter: []gzip.RequestFilter{
				gzip.NewCommonRequestFilter(),
				gzip.DefaultExtensionFilter(),
			},
			// ResponseHeaderFilter decide whether or not to compress response
			// judging by response header
			ResponseHeaderFilter: []gzip.ResponseHeaderFilter{
				gzip.NewSkipCompressedFilter(),
				gzip.DefaultContentTypeFilter(),
			},
		}).Gin)
	}

	const dist = "frontend-antd/dist/"

	// 404
	noRoute := func(ctx *gin.Context) {
		if strings.HasPrefix(ctx.GetHeader("Accept"), "text/html") {
			f, _ := os.ReadFile(dist + "404.html")
			ctx.Data(404, gin.MIMEHTML, f)
			return
		}
		ctx.Status(404)
		ctx.Writer.Write(nil)
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
			// 未登录，脚本重定向，防止客户端丢失缓存
			redirect(ctx, 401, "./login/")
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
				// 已登录，脚本重定向，防止客户端丢失缓存
				redirect(ctx, 401, "../")
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

var handlerGinLogger = gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
	mylog.UpdateWriter()
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
	} else {
		v += "\n"
	}
	return v
})

func handlerCheckNotLoggedIn401(ctx *gin.Context) {
	if !mysession.CheckLoggedInCookieForCtx(ctx) {
		ctx.AbortWithStatus(401)
	}
}

func handlerRemoveQuery(ctx *gin.Context) {
	if ctx.Request.URL.RawQuery != "" {
		// 移除参数
		redirect(ctx, 301, ctx.Request.URL.Path)
		ctx.Abort()
	}
}

func redirect(ctx *gin.Context, code int, path string) {
	if code/100 == 3 {
		hlfhr.Redirect(ctx.Writer, code, path)
	} else {
		ctx.Data(code, "text/html; charset=utf-8", []byte(`<script>location.replace(`+strconv.Quote(path)+`+location.hash)</script>`))
	}
}

func getClientIP(ctx *gin.Context) string {
	if conf.Http.Only_EnableXRealIp && getXForwarderAuth(ctx) {
		XRealIp := ctx.GetHeader("X-Real-Ip")
		if XRealIp != "" {
			return XRealIp
		}
	}
	return ctx.RemoteIP()
}

func getXForwarderAuth(ctx *gin.Context) bool {
	return conf.Http.Only_XForwarderAuth == "" || hmac.Equal([]byte(ctx.GetHeader("X-Forwarder-Auth")), []byte(conf.Http.Only_XForwarderAuth))
}
