package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jessewkun/gocommon/logger"
	"github.com/jessewkun/gocommon/response"
	"golang.org/x/time/rate"
)

// RateLimiterConfig 限流配置
type RateLimiterConfig struct {
	// 全局限流器，开启后，所有请求共享一个限流器，无论来自哪个IP
	GlobalLimiter *rate.Limiter
	// IP 限流器，开启后，为每个IP都创建一个限流器，单独限流
	IPLimiters map[string]*rate.Limiter
	// IP 限流配置，每个IP每秒允许的请求数
	IPLimit rate.Limit
	// IP 限流器最大积累的请求数
	IPBurst int
	// 是否启用 IP 限流，开启后，每个IP都受限流限制
	EnableIPLimit bool
	// 是否记录限流日志，开启后，会记录限流日志
	EnableLog bool
	// IP 限流器清理间隔，清理长时间未使用的 IP 限流器
	CleanupInterval time.Duration
	// IP 白名单，这些IP不受限流限制，开启后，这些IP不受限流限制
	Whitelist []string
	// 白名单IP是否也跳过全局限流，开启后，白名单IP也受全局限流限制
	WhitelistSkipGlobal bool
	// 自定义白名单检查函数，开启后，使用自定义白名单检查函数
	WhitelistChecker func(ip string) bool
}

// DefaultRateLimiterConfig 返回默认配置
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		GlobalLimiter:       rate.NewLimiter(rate.Limit(100), 200), // 每秒100个请求，最多积累200个
		IPLimiters:          make(map[string]*rate.Limiter),
		IPLimit:             rate.Limit(10), // 每个IP每秒10个请求
		IPBurst:             20,             // 每个IP最多积累20个请求
		EnableIPLimit:       true,
		EnableLog:           true,
		CleanupInterval:     time.Minute * 10,
		Whitelist:           []string{}, // 默认无白名单
		WhitelistSkipGlobal: false,      // 默认白名单IP也受全局限流影响
		WhitelistChecker:    nil,        // 默认使用内置检查逻辑
	}
}

var (
	config     *RateLimiterConfig
	configOnce sync.Once
	ipMutex    sync.RWMutex
)

// isWhitelisted 检查IP是否在白名单中
func isWhitelisted(ip string) bool {
	// 优先使用自定义检查函数
	if config.WhitelistChecker != nil {
		return config.WhitelistChecker(ip)
	}

	// 使用内置白名单检查
	if len(config.Whitelist) == 0 {
		return false
	}

	for _, whitelistIP := range config.Whitelist {
		if ip == whitelistIP {
			return true
		}
	}
	return false
}

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
		clientIP := c.ClientIP()
		isWhitelist := isWhitelisted(clientIP)

		// 检查全局限流（根据配置决定白名单IP是否跳过全局限流）
		if !config.WhitelistSkipGlobal || !isWhitelist {
			if !config.GlobalLimiter.Allow() {
				if config.EnableLog {
					logger.Warn(c.Request.Context(), TAG, "Global rate limit exceeded")
				}
				c.JSON(http.StatusOK, response.RateLimiterErrorResp(c))
				c.Abort()
				return
			}
		}

		// 检查 IP 限流（白名单IP跳过IP限流）
		if config.EnableIPLimit && !isWhitelist {
			limiter := getIPLimiter(clientIP)
			if !limiter.Allow() {
				if config.EnableLog {
					logger.Warn(c.Request.Context(), TAG, "IP rate limit exceeded: %s", clientIP)
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
		logger.Info(context.Background(), TAG, "Current IP limiters count: %d", len(config.IPLimiters))
	}
}
