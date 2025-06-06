package middleware

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/jessewkun/gocommon/constant"
	"github.com/jessewkun/gocommon/logger"

	"github.com/gin-gonic/gin"
)

// IOLog
// 返回结果前记录接口返回数据
func IOLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		bodyByte := []byte{}
		if c.Request.Method == http.MethodPost {
			bodyByte, _ = io.ReadAll(c.Request.Body)
		}
		var ctxResp any
		ctxResp, _ = c.Get(constant.CTX_API_OUTPUT)

		var fn func(c context.Context, tag string, msg string, field map[string]interface{})
		fn = logger.InfoWithField

		status := c.Writer.Status()
		if status >= http.StatusInternalServerError {
			fn = logger.ErrorWithField
		} else if status >= http.StatusBadRequest {
			fn = logger.WarnWithField
		}

		fn(c.Request.Context(), TAGNAME, "IOLOG", map[string]interface{}{
			"duration":        time.Since(t),
			"request_uri":     c.Request.RequestURI,
			"method":          c.Request.Method,
			"domain":          c.Request.Host,
			"remote_ip":       c.ClientIP(),
			"user_agent":      c.Request.UserAgent(),
			"status":          c.Writer.Status(),
			"response":        &ctxResp,
			"response_length": c.Writer.Size(),
			"body":            string(bodyByte),
		})
	}
}
