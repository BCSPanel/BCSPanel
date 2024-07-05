package httprouter

import (
	"github.com/bddjr/BCSPanel/src/server/conf"
	"github.com/gin-gonic/gin"
)

type apiColorScheme struct{}

func (a apiColorScheme) Init(g *gin.RouterGroup) {
	g.GET("/", func(ctx *gin.Context) {
		ctx.String(200, conf.ConfigConfig.ColorScheme)
	})
}
