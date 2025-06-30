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
	// IP 限流器最后使用时间记录
	IPLastUsed map[string]time.Time
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
	// IP 限流器过期时间，超过此时间未使用的限流器将被清理
	IPExpiration time.Duration
	// IP 白名单，这些IP不受限流限制，开启后，这些IP不受限流限制
	Whitelist []string
	// 白名单IP是否也跳过全局限流，开启后，白名单IP也受全局限流限制
	WhitelistSkipGlobal bool
	// 自定义白名单检查函数，开启后，使用自定义白名单检查函数
	WhitelistChecker func(ip string) bool
	// 互斥锁，保护IP限流器的并发访问
	ipMutex sync.RWMutex
	// 清理定时器
	cleanupTicker *time.Ticker
	// 停止清理的信号通道
	stopCleanup chan struct{}
}

// DefaultRateLimiterConfig 返回默认配置
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		GlobalLimiter:       rate.NewLimiter(rate.Limit(100), 200), // 每秒100个请求，最多积累200个
		IPLimiters:          make(map[string]*rate.Limiter),
		IPLastUsed:          make(map[string]time.Time),
		IPLimit:             rate.Limit(10), // 每个IP每秒10个请求
		IPBurst:             20,             // 每个IP最多积累20个请求
		EnableIPLimit:       true,
		EnableLog:           true,
		CleanupInterval:     time.Minute * 10,
		IPExpiration:        time.Minute * 10,
		Whitelist:           []string{}, // 默认无白名单
		WhitelistSkipGlobal: false,      // 默认白名单IP也受全局限流影响
		WhitelistChecker:    nil,        // 默认使用内置检查逻辑
		stopCleanup:         make(chan struct{}),
	}
}

// isWhitelisted 检查IP是否在白名单中
func isWhitelisted(cfg *RateLimiterConfig, ip string) bool {
	// 优先使用自定义检查函数
	if cfg.WhitelistChecker != nil {
		return cfg.WhitelistChecker(ip)
	}

	// 使用内置白名单检查
	if len(cfg.Whitelist) == 0 {
		return false
	}

	for _, whitelistIP := range cfg.Whitelist {
		if ip == whitelistIP {
			return true
		}
	}
	return false
}

// RateLimiter 返回一个限流中间件
func RateLimiter(cfg *RateLimiterConfig) gin.HandlerFunc {
	if cfg == nil {
		cfg = DefaultRateLimiterConfig()
	}

	// 添加配置验证日志
	if cfg.EnableLog {
		if cfg.GlobalLimiter == nil {
			logger.Warn(context.Background(), "RATE_LIMITER", "GlobalLimiter is nil, global rate limiting is disabled")
		} else {
			logger.Info(context.Background(), "RATE_LIMITER", "Rate limiter initialized with global limiter")
		}
	}

	// 启动 IP 限流器清理
	if cfg.EnableIPLimit {
		cfg.cleanupTicker = time.NewTicker(cfg.CleanupInterval)
		go func() {
			for {
				select {
				case <-cfg.cleanupTicker.C:
					cleanupIPLimiters(cfg)
				case <-cfg.stopCleanup:
					cfg.cleanupTicker.Stop()
					return
				}
			}
		}()
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		isWhitelist := isWhitelisted(cfg, clientIP)

		// 检查全局限流（根据配置决定白名单IP是否跳过全局限流）
		if (!cfg.WhitelistSkipGlobal || !isWhitelist) && cfg.GlobalLimiter != nil {
			if !cfg.GlobalLimiter.Allow() {
				if cfg.EnableLog {
					logger.Warn(c.Request.Context(), "RATE_LIMITER", "Global rate limit exceeded, url: %s", c.Request.URL.Path)
				}
				c.JSON(http.StatusOK, response.RateLimiterErrorResp(c))
				c.Abort()
				return
			}
		}

		// 检查 IP 限流（白名单IP跳过IP限流）
		if cfg.EnableIPLimit && !isWhitelist {
			limiter := getIPLimiter(cfg, clientIP)
			if limiter != nil && !limiter.Allow() {
				if cfg.EnableLog {
					logger.Warn(c.Request.Context(), "RATE_LIMITER", "IP rate limit exceeded: %s, url: %s", clientIP, c.Request.URL.Path)
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
func getIPLimiter(cfg *RateLimiterConfig, ip string) *rate.Limiter {
	cfg.ipMutex.RLock()
	limiter, exists := cfg.IPLimiters[ip]
	cfg.ipMutex.RUnlock()

	if exists {
		// 更新最后使用时间
		cfg.ipMutex.Lock()
		cfg.IPLastUsed[ip] = time.Now()
		cfg.ipMutex.Unlock()
		return limiter
	}

	cfg.ipMutex.Lock()
	defer cfg.ipMutex.Unlock()

	// 双重检查
	limiter, exists = cfg.IPLimiters[ip]
	if exists {
		// 更新最后使用时间
		cfg.IPLastUsed[ip] = time.Now()
		return limiter
	}

	limiter = rate.NewLimiter(cfg.IPLimit, cfg.IPBurst)
	cfg.IPLimiters[ip] = limiter
	cfg.IPLastUsed[ip] = time.Now()
	return limiter
}

// cleanupIPLimiters 清理长时间未使用的 IP 限流器
func cleanupIPLimiters(cfg *RateLimiterConfig) {
	cfg.ipMutex.Lock()
	defer cfg.ipMutex.Unlock()

	now := time.Now()
	expiredIPs := make([]string, 0)

	// 找出过期的IP限流器
	for ip, lastUsed := range cfg.IPLastUsed {
		if now.Sub(lastUsed) > cfg.IPExpiration {
			expiredIPs = append(expiredIPs, ip)
		}
	}

	// 删除过期的IP限流器
	for _, ip := range expiredIPs {
		delete(cfg.IPLimiters, ip)
		delete(cfg.IPLastUsed, ip)
	}

	// 记录清理日志
	if cfg.EnableLog {
		if len(expiredIPs) > 0 {
			logger.Info(context.Background(), "RATE_LIMITER", "Cleaned up %d expired IP limiters: %v", len(expiredIPs), expiredIPs)
		}
		logger.Info(context.Background(), "RATE_LIMITER", "Current IP limiters count: %d", len(cfg.IPLimiters))
	}
}

// RemoveIPLimiter 手动删除指定IP的限流器
func RemoveIPLimiter(cfg *RateLimiterConfig, ip string) bool {
	cfg.ipMutex.Lock()
	defer cfg.ipMutex.Unlock()

	if _, exists := cfg.IPLimiters[ip]; exists {
		delete(cfg.IPLimiters, ip)
		delete(cfg.IPLastUsed, ip)

		if cfg.EnableLog {
			logger.Info(context.Background(), "RATE_LIMITER", "Manually removed IP limiter for: %s", ip)
		}
		return true
	}

	return false
}

// GetIPLimitersCount 获取当前IP限流器数量
func GetIPLimitersCount(cfg *RateLimiterConfig) int {
	cfg.ipMutex.RLock()
	defer cfg.ipMutex.RUnlock()

	return len(cfg.IPLimiters)
}

// GetIPLimitersInfo 获取IP限流器详细信息
func GetIPLimitersInfo(cfg *RateLimiterConfig) map[string]time.Time {
	cfg.ipMutex.RLock()
	defer cfg.ipMutex.RUnlock()

	// 创建副本避免并发问题
	info := make(map[string]time.Time)
	for ip, lastUsed := range cfg.IPLastUsed {
		info[ip] = lastUsed
	}

	return info
}

// StopCleanup 停止清理定时器
func StopCleanup(cfg *RateLimiterConfig) {
	if cfg.cleanupTicker != nil {
		close(cfg.stopCleanup)
	}
}
