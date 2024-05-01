package httprouter

import (
	"github.com/bddjr/BCSPanel/src/server/conf"
	"github.com/gin-gonic/gin"
)

var ColorScheme = ""

func UpdateColorScheme() {
	switch conf.ConfigConfig.ColorScheme {
	case "light", "dark":
		ColorScheme = conf.ConfigConfig.ColorScheme
	default:
		ColorScheme = ""
	}
}

func colorSchemeHandler(ctx *gin.Context) {
	ctx.String(200, ColorScheme)
}
