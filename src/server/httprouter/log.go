package httprouter

import (
	"fmt"

	"github.com/bddjr/BCSPanel/src/server/conf"
	"github.com/bddjr/BCSPanel/src/server/mylog"
	"github.com/gin-gonic/gin"
)

var GinLoggerWithFormat = gin.LoggerWithFormatter(func(param gin.LogFormatterParams) (v string) {
	mylog.UpdateWriter()
	v = fmt.Sprint(
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
	}
	v += "\n"
	return
})

func GinLoggerHandler(ctx *gin.Context) {
	if conf.Http.Only_EnableGinLog || ctx.Errors.String() != "" {
		GinLoggerWithFormat(ctx)
	}
}
