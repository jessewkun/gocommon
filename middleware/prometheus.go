package middleware

import (
	"fmt"
	"time"

	"github.com/jessewkun/gocommon/prometheus"

	"github.com/gin-gonic/gin"
)

// Prometheus 收集和导出 Prometheus 指标
func Prometheus() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start).Seconds()
		statusCode := fmt.Sprintf("%d", c.Writer.Status())

		// 只记录注册过的路由（否则很多 path 会成为不同的 label）
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		prometheus.RequestsTotal.WithLabelValues(c.Request.Method, route, statusCode).Inc()
		prometheus.RequestDuration.WithLabelValues(c.Request.Method, route).Observe(duration)
	}
}
