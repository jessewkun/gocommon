package middleware

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jessewkun/gocommon/constant"
)

// Traceid
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceId := c.GetHeader(constant.CTX_TRACE_ID)
		if len(traceId) < 1 {
			traceId = uuid.New().String()
		}
		c.Set(constant.CTX_TRACE_ID, traceId)
		ctx := c.Request.Context()
		// 除了api接口层接受的是 gin.Context，其他地方都是 context.Context
		// 为了方便后续其他地方处理，比如后续代码逻辑获取 trace_id 或者日志默认打印 trace_id（config log transparent_parameter 配置中如果有），这里同步把 trace_id 放到 context.Context 中
		ctx = context.WithValue(ctx, constant.CTX_TRACE_ID, traceId)
		c.Request = c.Request.WithContext(ctx)
		host, _ := os.Hostname()
		c.Header("server", host)
		c.Next()
	}
}
