package httprouter

import (
	"os"

	"github.com/bddjr/BCSPanel/src/server/mylog"
	"github.com/gin-gonic/gin"
)

var _404html []byte

// var _401html []byte

func UPdate404html() {
	mylog.INFOln("http UPdate404html")
	var err error
	_404html, err = os.ReadFile("./src/404.html")
	if err != nil {
		mylog.ERRORln(err)
		_404html = []byte("404 Not Found\n")
	}
	// _401html, err = os.ReadFile("./src/401.html")
	// if err != nil {
	// 	mylog.ERRORln(err)
	// 	_401html = []byte("401 Unauthorized\n")
	// }
}

func _404Handler(ctx *gin.Context) {
	ctx.Writer.Header().Del("Cache-Control")
	ctx.Data(404, gin.MIMEHTML, _404html)
}

// func AbortWith401(ctx *gin.Context) {
// 	ctx.Writer.Header().Del("Cache-Control")
// 	ctx.Data(401, gin.MIMEHTML, _401html)
// 	ctx.Abort()
// }
