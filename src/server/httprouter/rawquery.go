package httprouter

import "github.com/gin-gonic/gin"

func GetRawQuery(RawQuery string) string {
	if RawQuery != "" {
		return "?" + RawQuery
	}
	return RawQuery
}

func GetRawQueryFromCtx(ctx *gin.Context) string {
	return GetRawQuery(ctx.Request.URL.RawQuery)
}
