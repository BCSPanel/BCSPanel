package routertools

import (
	"github.com/bddjr/BCSPanel/src/mysession"
	"github.com/bddjr/hlfhr"
	"github.com/gin-gonic/gin"
)

// 鉴权中间件，用于拒绝未登录的请求。
func CheckNotLoggedIn401(ctx *gin.Context) {
	if !mysession.CheckCtx(ctx) {
		ctx.AbortWithStatus(401)
	}
}

// 纯净的重定向，响应无body。
func Redirect(ctx *gin.Context, code int, path string) {
	hlfhr.Redirect(ctx.Writer, code, path)
}
