package httprouter

import (
	"crypto/hmac"

	"github.com/bddjr/BCSPanel/src/server/conf"
	"github.com/gin-gonic/gin"
)

func ClientIP(ctx *gin.Context) string {
	if conf.Http.Only_EnableXRealIp && XForwarderAuth(ctx) {
		XRealIp := ctx.GetHeader("X-Real-Ip")
		if XRealIp != "" {
			return XRealIp
		}
	}
	return ctx.RemoteIP()
}

func XForwarderAuth(ctx *gin.Context) bool {
	return conf.Http.Only_XForwarderAuth == "" || hmac.Equal([]byte(ctx.GetHeader("X-Forwarder-Auth")), []byte(conf.Http.Only_XForwarderAuth))
}
