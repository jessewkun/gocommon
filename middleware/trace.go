package middleware

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jessewkun/gocommon/constant"
)

// Trace 设置 trace_id
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(string(constant.CtxTraceID))
		if len(traceID) < 1 {
			traceID = uuid.New().String()
		}
		c.Set(string(constant.CtxTraceID), traceID)
		ctx := c.Request.Context()
		// 除了api接口层接受的是 gin.Context，其他地方都是 context.Context
		// 为了方便后续其他地方处理，比如后续代码逻辑获取 trace_id 或者日志默认打印 trace_id（config log transparent_parameter 配置中如果有），这里同步把 trace_id 放到 context.Context 中
		ctx = context.WithValue(ctx, constant.CtxTraceID, traceID)
		ctx = context.WithValue(ctx, constant.CtxCurrentRequestPath, c.Request.URL.Path)
		c.Request = c.Request.WithContext(ctx)
		host, _ := os.Hostname()
		c.Header("server", host)
		c.Next()
	}
}
