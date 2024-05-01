package httprouter

import (
	"github.com/bddjr/BCSPanel/src/server/conf"
	"github.com/bddjr/BCSPanel/src/server/mylog"
	"github.com/nanmu42/gzip"
)

// 更新Gzip压缩器
func UpdateGzipHandler() {
	if conf.Http.Old_GzipLevel == conf.Http.New_GzipLevel && conf.Http.Old_GzipMinContentLength == conf.Http.New_GzipMinContentLength {
		return
	}
	conf.Http.Old_GzipLevel = conf.Http.New_GzipLevel
	conf.Http.Old_GzipMinContentLength = conf.Http.New_GzipMinContentLength
	if conf.Http.Old_GzipLevel == 0 {
		mylog.INFOln("http Gzip off")
		GzipHandler = nil
		return
	}
	mylog.INFOf("http GzipLevel %d , GzipMinContentLength %d\n", conf.Http.Old_GzipLevel, conf.Http.Old_GzipMinContentLength)
	GzipHandler = gzip.NewHandler(gzip.Config{
		// gzip compression level to use
		CompressionLevel: int(conf.Http.Old_GzipLevel),
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
	}).Gin
}
