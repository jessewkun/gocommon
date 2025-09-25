// Package router 用于注册系统路由
package router

import (
	"net/http/pprof"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jessewkun/gocommon/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
)

// pprof 限流配置 - 仅允许本地访问
var pprofRateLimitConfig = &middleware.RateLimiterConfig{
	GlobalLimiter:       nil, // 关闭全局限流
	IPLimiters:          make(map[string]*rate.Limiter),
	IPLastUsed:          make(map[string]time.Time),
	IPLimit:             0, // 对非白名单IP设置为0，即完全禁止
	IPBurst:             0,
	EnableIPLimit:       true,
	EnableLog:           true,
	CleanupInterval:     time.Minute * 10,
	IPExpiration:        time.Minute * 10,
	Whitelist:           []string{"127.0.0.1", "::1", "localhost"}, // 本地IP白名单
	WhitelistSkipGlobal: true,                                      // 白名单IP跳过全局限流
	WhitelistChecker:    nil,                                       // 使用内置白名单检查
}

// RegisterSystemRoutes 注册系统路由，非必要勿动
func RegisterSystemRoutes(r *gin.Engine) {
	// ping
	r.GET("/health/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// prometheus metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// pprof 路由 - 仅允许本地访问
	pprofGroup := r.Group("/debug/pprof")
	pprofGroup.Use(middleware.RateLimiter(pprofRateLimitConfig))
	{
		pprofGroup.GET("/", gin.WrapF(pprof.Index))
		pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
		pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
		pprofGroup.GET("/allocs", gin.WrapF(pprof.Handler("allocs").ServeHTTP))
		pprofGroup.GET("/block", gin.WrapF(pprof.Handler("block").ServeHTTP))
		pprofGroup.GET("/goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
		pprofGroup.GET("/heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
		pprofGroup.GET("/mutex", gin.WrapF(pprof.Handler("mutex").ServeHTTP))
		pprofGroup.GET("/threadcreate", gin.WrapF(pprof.Handler("threadcreate").ServeHTTP))
	}
}
