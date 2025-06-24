package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jessewkun/gocommon/logger"
	"github.com/jessewkun/gocommon/response"
	"golang.org/x/time/rate"
)

// 限制每秒允许1次请求，最多积累3个请求
// var limiter = rate.NewLimiter(1, 3)

// RateLimiterConfig 限流配置
type RateLimiterConfig struct {
	// 全局限流器
	GlobalLimiter *rate.Limiter
	// IP 限流器
	IPLimiters map[string]*rate.Limiter
	// IP 限流配置
	IPLimit rate.Limit
	IPBurst int
	// 是否启用 IP 限流
	EnableIPLimit bool
	// 是否记录限流日志
	EnableLog bool
	// IP 限流器清理间隔
	CleanupInterval time.Duration
}

// DefaultRateLimiterConfig 返回默认配置
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		GlobalLimiter:   rate.NewLimiter(rate.Limit(100), 200), // 每秒100个请求，最多积累200个
		IPLimiters:      make(map[string]*rate.Limiter),
		IPLimit:         rate.Limit(10), // 每个IP每秒10个请求
		IPBurst:         20,             // 每个IP最多积累20个请求
		EnableIPLimit:   true,
		EnableLog:       true,
		CleanupInterval: time.Hour,
	}
}

var (
	config     *RateLimiterConfig
	configOnce sync.Once
	ipMutex    sync.RWMutex
)

// RateLimiter 返回一个限流中间件
func RateLimiter(cfg *RateLimiterConfig) gin.HandlerFunc {
	configOnce.Do(func() {
		if cfg == nil {
			cfg = DefaultRateLimiterConfig()
		}
		config = cfg

		// 启动 IP 限流器清理
		if cfg.EnableIPLimit {
			go func() {
				for {
					time.Sleep(cfg.CleanupInterval)
					cleanupIPLimiters()
				}
			}()
		}
	})

	return func(c *gin.Context) {
		// 检查全局限流
		if !config.GlobalLimiter.Allow() {
			if config.EnableLog {
				logger.Warn(c.Request.Context(), TAG, "Global rate limit exceeded")
			}
			c.JSON(http.StatusOK, response.RateLimiterErrorResp(c))
			c.Abort()
			return
		}

		// 检查 IP 限流
		if config.EnableIPLimit {
			ip := c.ClientIP()
			limiter := getIPLimiter(ip)
			if !limiter.Allow() {
				if config.EnableLog {
					logger.Warn(c.Request.Context(), TAG, "IP rate limit exceeded: %s", ip)
				}
				c.JSON(http.StatusOK, response.RateLimiterErrorResp(c))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// getIPLimiter 获取或创建 IP 限流器
func getIPLimiter(ip string) *rate.Limiter {
	ipMutex.RLock()
	limiter, exists := config.IPLimiters[ip]
	ipMutex.RUnlock()

	if exists {
		return limiter
	}

	ipMutex.Lock()
	defer ipMutex.Unlock()

	// 双重检查
	limiter, exists = config.IPLimiters[ip]
	if exists {
		return limiter
	}

	limiter = rate.NewLimiter(config.IPLimit, config.IPBurst)
	config.IPLimiters[ip] = limiter
	return limiter
}

// cleanupIPLimiters 清理长时间未使用的 IP 限流器
func cleanupIPLimiters() {
	ipMutex.Lock()
	defer ipMutex.Unlock()

	// 这里可以添加清理逻辑，比如删除长时间未使用的 IP 限流器
	// 为了简单起见，这里只是打印一下当前 IP 限流器的数量
	if config.EnableLog {
		logger.Info(nil, TAG, "Current IP limiters count: %d", len(config.IPLimiters))
	}
}
